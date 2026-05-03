package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintScaffoldNextMatchesStarterEntrypoint(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		want         []string
		notWant      []string
	}{
		{
			name:         "golayout",
			templateName: starterTemplateGoLayout,
			want:         []string{"go mod tidy", "go run ./cmd/app"},
		},
		{
			name:         "multi-flat-wire",
			templateName: starterTemplateMultiFlatWire,
			want:         []string{"go mod tidy", "make generate", "make run-user"},
			notWant:      []string{"go run ./cmd/app"},
		},
		{
			name:         "multi-independent",
			templateName: starterTemplateMultiIndependent,
			want:         []string{"go work sync", "make generate", "make run-user"},
			notWant:      []string{"go run ./cmd/app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			printScaffoldNext(buf, "example", tt.templateName)
			out := buf.String()
			for _, want := range tt.want {
				require.Contains(t, out, want)
			}
			for _, notWant := range tt.notWant {
				require.NotContains(t, out, notWant)
			}
		})
	}
}
