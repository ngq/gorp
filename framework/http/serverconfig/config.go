// Package serverconfig parses and validates HTTP service configuration.
package serverconfig

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

var serviceNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var (
	rootKeys = map[string]struct{}{
		"addr": {}, "mode": {}, "read_timeout": {}, "write_timeout": {},
		"idle_timeout": {}, "shutdown_timeout": {}, "services": {},
	}
	serviceKeys = map[string]struct{}{
		"enabled": {}, "addr": {}, "mode": {}, "read_timeout": {},
		"write_timeout": {}, "idle_timeout": {}, "shutdown_timeout": {},
	}
)

// Service describes one HTTP server instance.
type Service struct {
	Name            string
	Enabled         bool
	Addr            string
	Mode            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type rootConfig struct {
	Addr            string                        `mapstructure:"addr"`
	Mode            string                        `mapstructure:"mode"`
	ReadTimeout     string                        `mapstructure:"read_timeout"`
	WriteTimeout    string                        `mapstructure:"write_timeout"`
	IdleTimeout     string                        `mapstructure:"idle_timeout"`
	ShutdownTimeout string                        `mapstructure:"shutdown_timeout"`
	Services        map[string]namedServiceConfig `mapstructure:"services"`
}

type namedServiceConfig struct {
	Enabled         *bool  `mapstructure:"enabled"`
	Addr            string `mapstructure:"addr"`
	Mode            string `mapstructure:"mode"`
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	IdleTimeout     string `mapstructure:"idle_timeout"`
	ShutdownTimeout string `mapstructure:"shutdown_timeout"`
}

// Parse reads server.http. The single-service shorthand and named services
// form are intentionally mutually exclusive.
func Parse(cfg datacontract.Config) ([]Service, error) {
	if cfg == nil {
		return []Service{defaults(transportcontract.DefaultHTTPServiceName, ":8080")}, nil
	}

	var root rootConfig
	if err := cfg.Unmarshal("server.http", &root); err != nil {
		return nil, fmt.Errorf("unmarshal server.http: %w", err)
	}
	if err := validateKnownKeys(cfg.Get("server.http")); err != nil {
		return nil, err
	}

	legacyAddr := strings.TrimSpace(cfg.GetString("app.address"))
	hasSingle := strings.TrimSpace(root.Addr) != "" || legacyAddr != ""
	hasNamed := len(root.Services) > 0
	if hasSingle && hasNamed {
		return nil, fmt.Errorf("server.http.addr and server.http.services are mutually exclusive")
	}

	if !hasNamed {
		addr := strings.TrimSpace(root.Addr)
		if addr == "" {
			addr = legacyAddr
		}
		if addr == "" {
			return nil, fmt.Errorf("server.http.addr is required")
		}
		svc := defaults(transportcontract.DefaultHTTPServiceName, addr)
		svc.Mode = strings.TrimSpace(root.Mode)
		if err := validateMode(svc.Mode); err != nil {
			return nil, fmt.Errorf("server.http.mode: %w", err)
		}
		if err := applyDurations(&svc, root.ReadTimeout, root.WriteTimeout, root.IdleTimeout, root.ShutdownTimeout); err != nil {
			return nil, fmt.Errorf("server.http: %w", err)
		}
		applyLegacyTimeouts(cfg, &svc)
		return []Service{svc}, nil
	}

	names := make([]string, 0, len(root.Services))
	for name := range root.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]Service, 0, len(names))
	addresses := make(map[string]string, len(names))
	explicitMode := ""
	for _, name := range names {
		raw := root.Services[name]
		if !serviceNamePattern.MatchString(name) {
			return nil, fmt.Errorf("server.http.services.%s: invalid service name", name)
		}
		enabled := true
		if raw.Enabled != nil {
			enabled = *raw.Enabled
		}
		addr := strings.TrimSpace(raw.Addr)
		if enabled && addr == "" {
			return nil, fmt.Errorf("server.http.services.%s.addr is required when enabled", name)
		}
		if addr != "" {
			if previous, exists := addresses[addr]; exists {
				return nil, fmt.Errorf("server.http.services.%s.addr duplicates service %s: %s", name, previous, addr)
			}
			addresses[addr] = name
		}
		svc := defaults(name, addr)
		svc.Enabled = enabled
		svc.Mode = strings.TrimSpace(raw.Mode)
		if err := validateMode(svc.Mode); err != nil {
			return nil, fmt.Errorf("server.http.services.%s.mode: %w", name, err)
		}
		if svc.Mode != "" {
			if explicitMode != "" && explicitMode != svc.Mode {
				return nil, fmt.Errorf("server.http.services: Gin mode must be identical for all services")
			}
			explicitMode = svc.Mode
		}
		if err := applyDurations(&svc, raw.ReadTimeout, raw.WriteTimeout, raw.IdleTimeout, raw.ShutdownTimeout); err != nil {
			return nil, fmt.Errorf("server.http.services.%s: %w", name, err)
		}
		result = append(result, svc)
	}
	return result, nil
}

func validateMode(mode string) error {
	if mode == "" || mode == "debug" || mode == "release" || mode == "test" {
		return nil
	}
	return fmt.Errorf("must be one of debug, release, test")
}

func validateKnownKeys(raw any) error {
	root, ok := stringMap(raw)
	if !ok || root == nil {
		return nil
	}
	for key := range root {
		if _, known := rootKeys[strings.ToLower(key)]; !known {
			return fmt.Errorf("server.http.%s is unknown", key)
		}
	}
	services, ok := stringMap(root["services"])
	if !ok {
		return nil
	}
	for name, value := range services {
		fields, ok := stringMap(value)
		if !ok {
			continue
		}
		for key := range fields {
			if _, known := serviceKeys[strings.ToLower(key)]; !known {
				return fmt.Errorf("server.http.services.%s.%s is unknown", name, key)
			}
		}
	}
	return nil
}

func stringMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, value := range typed {
			text, ok := key.(string)
			if !ok {
				return nil, false
			}
			result[text] = value
		}
		return result, true
	default:
		return nil, false
	}
}

func defaults(name, addr string) Service {
	return Service{
		Name:            name,
		Enabled:         true,
		Addr:            addr,
		ReadTimeout:     defaultReadTimeout,
		WriteTimeout:    defaultWriteTimeout,
		IdleTimeout:     defaultIdleTimeout,
		ShutdownTimeout: defaultShutdownTimeout,
	}
}

func applyDurations(svc *Service, read, write, idle, shutdown string) error {
	values := []struct {
		name string
		raw  string
		dest *time.Duration
	}{
		{name: "read_timeout", raw: read, dest: &svc.ReadTimeout},
		{name: "write_timeout", raw: write, dest: &svc.WriteTimeout},
		{name: "idle_timeout", raw: idle, dest: &svc.IdleTimeout},
		{name: "shutdown_timeout", raw: shutdown, dest: &svc.ShutdownTimeout},
	}
	for _, value := range values {
		if strings.TrimSpace(value.raw) == "" {
			continue
		}
		parsed, err := time.ParseDuration(value.raw)
		if err != nil {
			return fmt.Errorf("%s is invalid: %w", value.name, err)
		}
		if parsed <= 0 {
			return fmt.Errorf("%s must be greater than zero", value.name)
		}
		*value.dest = parsed
	}
	return nil
}

func applyLegacyTimeouts(cfg datacontract.Config, svc *Service) {
	if value := cfg.GetInt("app.http.read_timeout_sec"); value > 0 {
		svc.ReadTimeout = time.Duration(value) * time.Second
	}
	if value := cfg.GetInt("app.http.write_timeout_sec"); value > 0 {
		svc.WriteTimeout = time.Duration(value) * time.Second
	}
	if value := cfg.GetInt("app.http.idle_timeout_sec"); value > 0 {
		svc.IdleTimeout = time.Duration(value) * time.Second
	}
}
