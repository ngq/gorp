// Package httpx provides OpenAPI 3.0 route introspection and auto-export capabilities.
package httpx

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// OpenAPI3Document represents a lightweight OpenAPI 3.0 root document.
type OpenAPI3Document struct {
	OpenAPI    string                                  `json:"openapi"`
	Info       OpenAPI3Info                            `json:"info"`
	Paths      map[string]map[string]OpenAPI3Operation `json:"paths"`
	Components *OpenAPI3Components                     `json:"components,omitempty"`
	Security   []map[string][]string                   `json:"security,omitempty"`
}

// OpenAPI3Components defines reusable OpenAPI schemas.
type OpenAPI3Components struct {
	SecuritySchemes map[string]OpenAPI3SecurityScheme `json:"securitySchemes,omitempty"`
}

// OpenAPI3SecurityScheme defines security schemes such as Bearer JWT.
type OpenAPI3SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// OpenAPI3Info describes API metadata.
type OpenAPI3Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// OpenAPI3Operation describes an HTTP operation on a path.
type OpenAPI3Operation struct {
	Summary    string                      `json:"summary"`
	Tags       []string                    `json:"tags,omitempty"`
	Parameters []OpenAPI3Parameter         `json:"parameters,omitempty"`
	Responses  map[string]OpenAPI3Response `json:"responses"`
}

// OpenAPI3Parameter describes a request parameter.
type OpenAPI3Parameter struct {
	Name     string            `json:"name"`
	In       string            `json:"in"` // path, query, header
	Required bool              `json:"required"`
	Schema   map[string]string `json:"schema"`
}

// OpenAPI3Response describes an API response.
type OpenAPI3Response struct {
	Description string `json:"description"`
}

var paramRegexp = regexp.MustCompile(`:([a-zA-Z0-9_]+)`)

// ExportOpenAPI3 inspects the Gin engine's routes and auto-generates an OpenAPI 3.0 spec document.
func ExportOpenAPI3(engine *gin.Engine, title, version string) *OpenAPI3Document {
	if title == "" {
		title = "gorp Auto-Exported API Spec"
	}
	if version == "" {
		version = "1.0.0"
	}

	doc := &OpenAPI3Document{
		OpenAPI: "3.0.0",
		Info: OpenAPI3Info{
			Title:       title,
			Version:     version,
			Description: "Auto-generated OpenAPI 3.0 spec from Gin route registry",
		},
		Paths: make(map[string]map[string]OpenAPI3Operation),
		Components: &OpenAPI3Components{
			SecuritySchemes: map[string]OpenAPI3SecurityScheme{
				"BearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
		},
		Security: []map[string][]string{
			{"BearerAuth": []string{}},
		},
	}

	if engine == nil {
		return doc
	}

	routes := engine.Routes()
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/debug/") {
			continue // Skip debug dashboard internal endpoints
		}

		// Convert Gin path parameters :id -> {id}
		openAPIPath := paramRegexp.ReplaceAllString(route.Path, "{$1}")

		method := strings.ToLower(route.Method)
		if method == "head" || method == "options" {
			continue
		}

		if doc.Paths[openAPIPath] == nil {
			doc.Paths[openAPIPath] = make(map[string]OpenAPI3Operation)
		}

		// Infer tags from first segment of path (e.g. /api/v1/users -> users)
		tags := []string{}
		segments := strings.Split(strings.Trim(route.Path, "/"), "/")
		if len(segments) > 0 && segments[0] != "" {
			tag := segments[0]
			if (tag == "api" || tag == "v1" || tag == "v2") && len(segments) > 1 {
				tag = segments[1]
			}
			tags = append(tags, tag)
		}

		// Extract path parameters
		params := []OpenAPI3Parameter{}
		matches := paramRegexp.FindAllStringSubmatch(route.Path, -1)
		for _, m := range matches {
			if len(m) > 1 {
				params = append(params, OpenAPI3Parameter{
					Name:     m[1],
					In:       "path",
					Required: true,
					Schema:   map[string]string{"type": "string"},
				})
			}
		}

		op := OpenAPI3Operation{
			Summary:    strings.ToUpper(method) + " " + route.Path,
			Tags:       tags,
			Parameters: params,
			Responses: map[string]OpenAPI3Response{
				"200": {Description: "Successful response"},
				"400": {Description: "Bad request"},
				"500": {Description: "Internal server error"},
			},
		}

		doc.Paths[openAPIPath][method] = op
	}

	return doc
}

// ExportOpenAPI3JSON converts the Gin routes to formatted OpenAPI 3.0 JSON.
func ExportOpenAPI3JSON(engine *gin.Engine, title, version string) string {
	doc := ExportOpenAPI3(engine, title, version)
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
