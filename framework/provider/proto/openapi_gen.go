// Package proto provides OpenAPI documentation generator implementation.
// This file implements OpenAPI 3.0 spec generation from proto files.
// Parses google.api.http annotations and generates REST API documentation.
//
// Proto 包提供 OpenAPI 文档生成器实现。
// 本文件实现从 proto 文件生成 OpenAPI 3.0 规范。
// 解析 google.api.http 注解并生成 REST API 文档。
package proto

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// GenOpenAPI generates OpenAPI/Swagger documentation from proto file.
// Parses proto annotations (google.api.http) and generates OpenAPI 3.0 spec.
//
// GenOpenAPI 从 proto 文件生成 OpenAPI/Swagger 文档。
// 解析 proto 注解 (google.api.http) 并生成 OpenAPI 3.0 规范。
func (g *Generator) GenOpenAPI(ctx context.Context, opts integrationcontract.OpenAPIGenOptions) error {
	if opts.ProtoFile == "" {
		return errors.New("proto file path is required")
	}
	if opts.OutputFile == "" {
		return errors.New("output file path is required")
	}

	// 解析 proto 文件内容（包含 message、enum、service）。
	protoContent, err := parseProtoFileContent(opts.ProtoFile)
	if err != nil {
		return fmt.Errorf("parse proto file content failed: %w", err)
	}

	// 如果指定了服务名则过滤。
	if opts.ServiceName != "" {
		filtered := []ProtoService{}
		for _, svc := range protoContent.Services {
			if svc.Name == opts.ServiceName {
				filtered = append(filtered, svc)
			}
		}
		protoContent.Services = filtered
	}

	if len(protoContent.Services) == 0 {
		return fmt.Errorf("no services found in proto file: %s", opts.ProtoFile)
	}

	// 生成 OpenAPI 文档（包含完整 Schema）。
	doc := generateOpenAPIDocWithSchemas(protoContent, opts)

	// 写入文件。
	outputDir := filepath.Dir(opts.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory failed: %w", err)
	}

	// 根据文件扩展名选择格式。
	var content string
	if strings.HasSuffix(strings.ToLower(opts.OutputFile), ".json") {
		content = openAPIToJSON(doc)
	} else {
		content = openAPIToYAML(doc)
	}

	return os.WriteFile(opts.OutputFile, []byte(content), 0644)
}

// OpenAPIDoc represents OpenAPI 3.0 document structure.
type OpenAPIDoc struct {
	OpenAPI    string              `json:"openapi" yaml:"openapi"`
	Info       OpenAPIInfo         `json:"info" yaml:"info"`
	Servers    []OpenAPIServer     `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components OpenAPIComponents   `json:"components,omitempty" yaml:"components,omitempty"`
}

// OpenAPIInfo represents OpenAPI info section.
type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

// OpenAPIServer represents OpenAPI server section.
type OpenAPIServer struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// PathItem represents OpenAPI path item.
type PathItem struct {
	Get     *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post    *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put     *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Summary string     `json:"summary,omitempty" yaml:"summary,omitempty"`
}

// Operation represents OpenAPI operation.
type Operation struct {
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string              `json:"operationId" yaml:"operationId"`
	Tags        []string            `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses" yaml:"responses"`
	Security    []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
}

// Parameter represents OpenAPI parameter.
type Parameter struct {
	Name        string `json:"name" yaml:"name"`
	In          string `json:"in" yaml:"in"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      Schema `json:"schema" yaml:"schema"`
}

// RequestBody represents OpenAPI request body.
type RequestBody struct {
	Required bool              `json:"required,omitempty" yaml:"required,omitempty"`
	Content  map[string]Media  `json:"content" yaml:"content"`
}

// Media represents OpenAPI media type.
type Media struct {
	Schema Schema `json:"schema" yaml:"schema"`
}

// Response represents OpenAPI response.
type Response struct {
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]Media   `json:"content,omitempty" yaml:"content,omitempty"`
}

// Schema represents OpenAPI schema.
type Schema struct {
	Type                 string            `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string            `json:"format,omitempty" yaml:"format,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Ref                  string            `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Items                *Schema           `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Enum                 []string          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Description          string            `json:"description,omitempty" yaml:"description,omitempty"`
	Example              interface{}       `json:"example,omitempty" yaml:"example,omitempty"`
	Required             []string          `json:"required,omitempty" yaml:"required,omitempty"`
}

// OpenAPIComponents represents OpenAPI components section.
type OpenAPIComponents struct {
	Schemas         map[string]Schema          `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme  `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// SecurityScheme represents OpenAPI security scheme.
type SecurityScheme struct {
	Type        string `json:"type" yaml:"type"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	In          string `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme      string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
}

// generateOpenAPIDoc generates OpenAPI document from proto services.
func generateOpenAPIDoc(services []ProtoService, opts integrationcontract.OpenAPIGenOptions) OpenAPIDoc {
	doc := OpenAPIDoc{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       opts.Title,
			Description: opts.Description,
			Version:     opts.Version,
		},
		Paths: make(map[string]PathItem),
		Components: OpenAPIComponents{
			Schemas:         make(map[string]Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// 设置默认值。
	if doc.Info.Title == "" {
		doc.Info.Title = "API"
	}
	if doc.Info.Version == "" {
		doc.Info.Version = "1.0.0"
	}

	// 添加服务器。
	if opts.Host != "" {
		scheme := "http"
		if len(opts.Schemes) > 0 {
			scheme = opts.Schemes[0]
		}
		doc.Servers = append(doc.Servers, OpenAPIServer{
			URL: fmt.Sprintf("%s://%s%s", scheme, opts.Host, opts.BasePath),
		})
	}

	// 为每个服务生成路径。
	for _, svc := range services {
		for _, m := range svc.Methods {
			if m.HTTPMethod == "" || m.HTTPPath == "" {
				continue
			}

			// 创建操作。
			op := &Operation{
				OperationID: fmt.Sprintf("%s_%s", svc.Name, m.Name),
				Summary:     m.Name,
				Tags:        []string{svc.Name},
				Responses: map[string]Response{
					"200": {
						Description: "Success",
						Content: map[string]Media{
							"application/json": {Schema: Schema{Type: "object"}},
						},
					},
					"400": {
						Description: "Bad Request",
						Content: map[string]Media{
							"application/json": {Schema: Schema{Type: "object"}},
						},
					},
					"500": {
						Description: "Internal Server Error",
						Content: map[string]Media{
							"application/json": {Schema: Schema{Type: "object"}},
						},
					},
				},
			}

			// 添加请求体（POST/PUT/PATCH）。
			if m.HTTPMethod == "POST" || m.HTTPMethod == "PUT" || m.HTTPMethod == "PATCH" {
				op.RequestBody = &RequestBody{
					Required: true,
					Content: map[string]Media{
						"application/json": {Schema: Schema{
							Type: "object",
						}},
					},
				}
			}

			// 解析路径参数。
			params := extractPathParams(m.HTTPPath)
			for _, p := range params {
				op.Parameters = append(op.Parameters, Parameter{
					Name:     p,
					In:       "path",
					Required: true,
					Schema:   Schema{Type: "string"},
				})
			}

			// 添加认证要求。
			if m.AuthRequired || len(m.AuthRoles) > 0 {
				op.Security = []map[string][]string{
					{"bearerAuth": {}},
				}
			}

			// 添加到路径。
			pathItem, exists := doc.Paths[m.HTTPPath]
			if !exists {
				pathItem = PathItem{}
			}

			switch m.HTTPMethod {
			case "GET":
				pathItem.Get = op
			case "POST":
				pathItem.Post = op
			case "PUT":
				pathItem.Put = op
			case "DELETE":
				pathItem.Delete = op
			case "PATCH":
				pathItem.Patch = op
			}

			doc.Paths[m.HTTPPath] = pathItem
		}
	}

	// 添加默认认证方案。
	doc.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}

	return doc
}

// generateOpenAPIDocWithSchemas generates OpenAPI document with full schemas from proto content.
// Includes message definitions in components/schemas section.
//
// generateOpenAPIDocWithSchemas 从 proto 内容生成包含完整 schema 的 OpenAPI 文档。
// 在 components/schemas 部分包含 message 定义。
func generateOpenAPIDocWithSchemas(protoContent *ProtoFileContent, opts integrationcontract.OpenAPIGenOptions) OpenAPIDoc {
	doc := OpenAPIDoc{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       opts.Title,
			Description: opts.Description,
			Version:     opts.Version,
		},
		Paths: make(map[string]PathItem),
		Components: OpenAPIComponents{
			Schemas:         make(map[string]Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// 设置默认值。
	if doc.Info.Title == "" {
		doc.Info.Title = "API"
	}
	if doc.Info.Version == "" {
		doc.Info.Version = "1.0.0"
	}

	// 添加服务器。
	if opts.Host != "" {
		scheme := "http"
		if len(opts.Schemes) > 0 {
			scheme = opts.Schemes[0]
		}
		doc.Servers = append(doc.Servers, OpenAPIServer{
			URL: fmt.Sprintf("%s://%s%s", scheme, opts.Host, opts.BasePath),
		})
	}

	// 生成 message schemas 到 components。
	for _, msg := range protoContent.Messages {
		schema := messageToSchema(msg, protoContent.Messages, protoContent.Enums)
		schema.Description = msg.Description
		doc.Components.Schemas[msg.Name] = schema
	}

	// 生成 enum schemas 到 components。
	for _, enum := range protoContent.Enums {
		enumSchema := enumToSchema(enum.Name, protoContent.Enums)
		enumSchema.Description = enum.Description
		doc.Components.Schemas[enum.Name] = enumSchema
	}

	// 为每个服务生成路径。
	for _, svc := range protoContent.Services {
		for _, m := range svc.Methods {
			if m.HTTPMethod == "" || m.HTTPPath == "" {
				continue
			}

			// 创建操作。
			op := &Operation{
				OperationID: fmt.Sprintf("%s_%s", svc.Name, m.Name),
				Summary:     m.Name,
				Tags:        []string{svc.Name},
				Responses: map[string]Response{
					"200": {
						Description: "Success",
						Content: map[string]Media{
							"application/json": {Schema: createSchemaRef(m.ResponseType, protoContent)},
						},
					},
					"400": {
						Description: "Bad Request",
						Content: map[string]Media{
							"application/json": {Schema: Schema{Type: "object"}},
						},
					},
					"500": {
						Description: "Internal Server Error",
						Content: map[string]Media{
							"application/json": {Schema: Schema{Type: "object"}},
						},
					},
				},
			}

			// 添加请求体（POST/PUT/PATCH）。
			if m.HTTPMethod == "POST" || m.HTTPMethod == "PUT" || m.HTTPMethod == "PATCH" {
				reqSchema := createSchemaRef(m.RequestType, protoContent)
				op.RequestBody = &RequestBody{
					Required: true,
					Content: map[string]Media{
						"application/json": {Schema: reqSchema},
					},
				}
			}

			// 解析路径参数。
			params := extractPathParams(m.HTTPPath)
			for _, p := range params {
				op.Parameters = append(op.Parameters, Parameter{
					Name:        p,
					In:          "path",
					Required:    true,
					Description: fmt.Sprintf("%s parameter", p),
					Schema:      Schema{Type: "string"},
				})
			}

			// 添加认证要求。
			if m.AuthRequired || len(m.AuthRoles) > 0 {
				op.Security = []map[string][]string{
					{"bearerAuth": {}},
				}
			}

			// 添加到路径。
			pathItem, exists := doc.Paths[m.HTTPPath]
			if !exists {
				pathItem = PathItem{}
			}

			switch m.HTTPMethod {
			case "GET":
				pathItem.Get = op
			case "POST":
				pathItem.Post = op
			case "PUT":
				pathItem.Put = op
			case "DELETE":
				pathItem.Delete = op
			case "PATCH":
				pathItem.Patch = op
			}

			doc.Paths[m.HTTPPath] = pathItem
		}
	}

	// 添加默认认证方案。
	doc.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}

	return doc
}

// createSchemaRef 创建 schema 引用，如果类型存在则使用 $ref，否则返回基本类型。
func createSchemaRef(typeName string, protoContent *ProtoFileContent) Schema {
	// 检查是否是已知 message
	for _, msg := range protoContent.Messages {
		if msg.Name == typeName {
			return Schema{Ref: "#/components/schemas/" + typeName}
		}
	}

	// 检查是否是已知 enum
	for _, enum := range protoContent.Enums {
		if enum.Name == typeName {
			return Schema{Ref: "#/components/schemas/" + typeName}
		}
	}

	// 检查特殊类型
	openapiType, format := protoTypeToOpenAPIType(typeName)
	if openapiType != "object" || format != "" {
		return Schema{Type: openapiType, Format: format}
	}

	// 未知类型，返回 object
	return Schema{Type: "object"}
}

// extractPathParams extracts path parameters from proto HTTP path.
// /v1/users/{id} -> ["id"]
func extractPathParams(path string) []string {
	var params []string
	start := -1
	for i, c := range path {
		if c == '{' {
			start = i + 1
		} else if c == '}' && start >= 0 {
			params = append(params, path[start:i])
			start = -1
		}
	}
	return params
}

// openAPIToYAML converts OpenAPI doc to YAML format.
func openAPIToYAML(doc OpenAPIDoc) string {
	var buf strings.Builder

	buf.WriteString("openapi: ")
	buf.WriteString(doc.OpenAPI)
	buf.WriteString("\n")

	buf.WriteString("info:\n")
	buf.WriteString("  title: ")
	buf.WriteString(doc.Info.Title)
	buf.WriteString("\n")
	if doc.Info.Description != "" {
		buf.WriteString("  description: ")
		buf.WriteString(doc.Info.Description)
		buf.WriteString("\n")
	}
	buf.WriteString("  version: ")
	buf.WriteString(doc.Info.Version)
	buf.WriteString("\n")

	if len(doc.Servers) > 0 {
		buf.WriteString("servers:\n")
		for _, s := range doc.Servers {
			buf.WriteString("  - url: ")
			buf.WriteString(s.URL)
			buf.WriteString("\n")
		}
	}

	buf.WriteString("paths:\n")
	for path, item := range doc.Paths {
		buf.WriteString("  ")
		buf.WriteString(path)
		buf.WriteString(":\n")

		if item.Get != nil {
			writeOperationYAML(&buf, "get", item.Get)
		}
		if item.Post != nil {
			writeOperationYAML(&buf, "post", item.Post)
		}
		if item.Put != nil {
			writeOperationYAML(&buf, "put", item.Put)
		}
		if item.Delete != nil {
			writeOperationYAML(&buf, "delete", item.Delete)
		}
		if item.Patch != nil {
			writeOperationYAML(&buf, "patch", item.Patch)
		}
	}

	buf.WriteString("components:\n")
	buf.WriteString("  securitySchemes:\n")
	buf.WriteString("    bearerAuth:\n")
	buf.WriteString("      type: http\n")
	buf.WriteString("      scheme: bearer\n")
	buf.WriteString("      bearerFormat: JWT\n")

	return buf.String()
}

// writeOperationYAML writes operation in YAML format.
func writeOperationYAML(buf *strings.Builder, method string, op *Operation) {
	buf.WriteString("    ")
	buf.WriteString(method)
	buf.WriteString(":\n")
	buf.WriteString("      operationId: ")
	buf.WriteString(op.OperationID)
	buf.WriteString("\n")
	if op.Summary != "" {
		buf.WriteString("      summary: ")
		buf.WriteString(op.Summary)
		buf.WriteString("\n")
	}
	if len(op.Tags) > 0 {
		buf.WriteString("      tags:\n")
		for _, t := range op.Tags {
			buf.WriteString("        - ")
			buf.WriteString(t)
			buf.WriteString("\n")
		}
	}
	if len(op.Parameters) > 0 {
		buf.WriteString("      parameters:\n")
		for _, p := range op.Parameters {
			buf.WriteString("        - name: ")
			buf.WriteString(p.Name)
			buf.WriteString("\n")
			buf.WriteString("          in: ")
			buf.WriteString(p.In)
			buf.WriteString("\n")
			buf.WriteString("          required: true\n")
			buf.WriteString("          schema:\n")
			buf.WriteString("            type: string\n")
		}
	}
	if op.RequestBody != nil {
		buf.WriteString("      requestBody:\n")
		buf.WriteString("        required: true\n")
		buf.WriteString("        content:\n")
		buf.WriteString("          application/json:\n")
		buf.WriteString("            schema:\n")
		buf.WriteString("              type: object\n")
	}
	buf.WriteString("      responses:\n")
	buf.WriteString("        '200':\n")
	buf.WriteString("          description: Success\n")
	buf.WriteString("        '400':\n")
	buf.WriteString("          description: Bad Request\n")
	buf.WriteString("        '500':\n")
	buf.WriteString("          description: Internal Server Error\n")
	if len(op.Security) > 0 {
		buf.WriteString("      security:\n")
		buf.WriteString("        - bearerAuth: []\n")
	}
}

// openAPIToJSON converts OpenAPI doc to JSON format.
func openAPIToJSON(doc OpenAPIDoc) string {
	var buf strings.Builder

	buf.WriteString("{\n")
	buf.WriteString("  \"openapi\": \"")
	buf.WriteString(doc.OpenAPI)
	buf.WriteString("\",\n")

	buf.WriteString("  \"info\": {\n")
	buf.WriteString("    \"title\": \"")
	buf.WriteString(doc.Info.Title)
	buf.WriteString("\",\n")
	if doc.Info.Description != "" {
		buf.WriteString("    \"description\": \"")
		buf.WriteString(doc.Info.Description)
		buf.WriteString("\",\n")
	}
	buf.WriteString("    \"version\": \"")
	buf.WriteString(doc.Info.Version)
	buf.WriteString("\"\n")
	buf.WriteString("  },\n")

	buf.WriteString("  \"paths\": {\n")
	first := true
	for path, item := range doc.Paths {
		if !first {
			buf.WriteString(",\n")
		}
		first = false

		buf.WriteString("    \"")
		buf.WriteString(path)
		buf.WriteString("\": {\n")

		if item.Get != nil {
			writeOperationJSON(&buf, "get", item.Get, true)
		}
		if item.Post != nil {
			if item.Get != nil {
				buf.WriteString(",\n")
			}
			writeOperationJSON(&buf, "post", item.Post, item.Get == nil)
		}
		buf.WriteString("\n    }")
	}
	buf.WriteString("\n  },\n")

	buf.WriteString("  \"components\": {\n")
	buf.WriteString("    \"securitySchemes\": {\n")
	buf.WriteString("      \"bearerAuth\": {\n")
	buf.WriteString("        \"type\": \"http\",\n")
	buf.WriteString("        \"scheme\": \"bearer\",\n")
	buf.WriteString("        \"bearerFormat\": \"JWT\"\n")
	buf.WriteString("      }\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")

	buf.WriteString("}\n")

	return buf.String()
}

// writeOperationJSON writes operation in JSON format.
func writeOperationJSON(buf *strings.Builder, method string, op *Operation, isFirst bool) {
	buf.WriteString("      \"")
	buf.WriteString(method)
	buf.WriteString("\": {\n")
	buf.WriteString("        \"operationId\": \"")
	buf.WriteString(op.OperationID)
	buf.WriteString("\"")
	if op.Summary != "" {
		buf.WriteString(",\n        \"summary\": \"")
		buf.WriteString(op.Summary)
		buf.WriteString("\"")
	}
	if len(op.Tags) > 0 {
		buf.WriteString(",\n        \"tags\": [")
		for i, t := range op.Tags {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("\"")
			buf.WriteString(t)
			buf.WriteString("\"")
		}
		buf.WriteString("]")
	}
	buf.WriteString(",\n        \"responses\": {\n")
	buf.WriteString("          \"200\": { \"description\": \"Success\" },\n")
	buf.WriteString("          \"400\": { \"description\": \"Bad Request\" },\n")
	buf.WriteString("          \"500\": { \"description\": \"Internal Server Error\" }\n")
	buf.WriteString("        }")
	if len(op.Security) > 0 {
		buf.WriteString(",\n        \"security\": [{ \"bearerAuth\": [] }]")
	}
	buf.WriteString("\n      }")
}
