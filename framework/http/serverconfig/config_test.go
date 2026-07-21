package serverconfig

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/spf13/viper"
)

type testConfig struct{ v *viper.Viper }

func configFromYAML(t *testing.T, content string) *testConfig {
	t.Helper()
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(content)); err != nil {
		t.Fatal(err)
	}
	return &testConfig{v: v}
}

func (c *testConfig) Env() string                 { return "test" }
func (c *testConfig) Get(key string) any          { return c.v.Get(key) }
func (c *testConfig) GetString(key string) string { return c.v.GetString(key) }
func (c *testConfig) GetInt(key string) int       { return c.v.GetInt(key) }
func (c *testConfig) GetBool(key string) bool     { return c.v.GetBool(key) }
func (c *testConfig) GetFloat(key string) float64 { return c.v.GetFloat64(key) }
func (c *testConfig) Unmarshal(key string, out any) error {
	return c.v.UnmarshalKey(key, out)
}
func (c *testConfig) Watch(context.Context, string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (c *testConfig) Reload(context.Context) error { return nil }

func TestParseSingleServiceShorthand(t *testing.T) {
	services, err := Parse(configFromYAML(t, `
server:
  http:
    addr: ":8080"
    read_timeout: 2s
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Name != "default" || services[0].Addr != ":8080" {
		t.Fatalf("unexpected services: %#v", services)
	}
	if services[0].ReadTimeout != 2*time.Second || !services[0].Enabled {
		t.Fatalf("unexpected single service defaults: %#v", services[0])
	}
}

func TestParseNamedServices(t *testing.T) {
	services, err := Parse(configFromYAML(t, `
server:
  http:
    services:
      wecom:
        enabled: false
        addr: ":8082"
      admin:
        addr: ":8081"
        shutdown_timeout: 3s
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 2 || services[0].Name != "admin" || services[1].Name != "wecom" {
		t.Fatalf("services must be stable and sorted: %#v", services)
	}
	if services[0].ShutdownTimeout != 3*time.Second || services[1].Enabled {
		t.Fatalf("unexpected named config: %#v", services)
	}
}

func TestParseRejectsAmbiguousAndInvalidConfigurations(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		message string
	}{
		{
			name:    "single and named",
			yaml:    `server: {http: {addr: ":8080", services: {admin: {addr: ":8081"}}}}`,
			message: "mutually exclusive",
		},
		{
			name:    "duplicate address",
			yaml:    `server: {http: {services: {admin: {addr: ":8080"}, wecom: {addr: ":8080"}}}}`,
			message: "duplicates service",
		},
		{
			name:    "invalid duration",
			yaml:    `server: {http: {addr: ":8080", read_timeout: "soon"}}`,
			message: "read_timeout is invalid",
		},
		{
			name:    "non-positive duration",
			yaml:    `server: {http: {addr: ":8080", idle_timeout: "0s"}}`,
			message: "must be greater than zero",
		},
		{
			name:    "unknown field",
			yaml:    `server: {http: {addr: ":8080", mystery: true}}`,
			message: "mystery is unknown",
		},
		{
			name:    "mixed gin modes",
			yaml:    `server: {http: {services: {admin: {addr: ":8081", mode: debug}, wecom: {addr: ":8082", mode: release}}}}`,
			message: "mode must be identical",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(configFromYAML(t, tt.yaml))
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("expected %q error, got %v", tt.message, err)
			}
		})
	}
}

func TestParseAllowsDisabledServiceWithoutAddress(t *testing.T) {
	services, err := Parse(configFromYAML(t, `
server:
  http:
    services:
      admin:
        addr: ":8081"
      wecom:
        enabled: false
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 2 || services[1].Enabled || services[1].Addr != "" {
		t.Fatalf("unexpected disabled service: %#v", services)
	}
}
