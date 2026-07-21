package gin

import (
	"sort"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type httpRegistry struct {
	entries map[string]transportcontract.HTTPServiceEntry
	names   []string
}

func newHTTPRegistry(entries []transportcontract.HTTPServiceEntry) *httpRegistry {
	registry := &httpRegistry{entries: make(map[string]transportcontract.HTTPServiceEntry, len(entries))}
	for _, entry := range entries {
		registry.entries[entry.Name] = entry
		registry.names = append(registry.names, entry.Name)
	}
	sort.Strings(registry.names)
	return registry
}

func (r *httpRegistry) Get(name string) (transportcontract.HTTP, bool) {
	if r == nil {
		return nil, false
	}
	entry, ok := r.entries[name]
	if !ok {
		return nil, false
	}
	return entry.Service, true
}

func (r *httpRegistry) Default() (transportcontract.HTTP, bool) {
	return r.Get(transportcontract.DefaultHTTPServiceName)
}

func (r *httpRegistry) Names() []string {
	if r == nil {
		return []string{}
	}
	return append([]string(nil), r.names...)
}

func (r *httpRegistry) Entries() []transportcontract.HTTPServiceEntry {
	if r == nil {
		return []transportcontract.HTTPServiceEntry{}
	}
	entries := make([]transportcontract.HTTPServiceEntry, 0, len(r.names))
	for _, name := range r.names {
		entries = append(entries, r.entries[name])
	}
	return entries
}

func (r *httpRegistry) Use(middleware ...transportcontract.Middleware) {
	if r == nil || len(middleware) == 0 {
		return
	}
	for _, name := range r.names {
		entry := r.entries[name]
		if entry.Service != nil {
			entry.Service.Use(middleware...)
		}
	}
}
