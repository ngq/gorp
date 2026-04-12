package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: verify-template <templates-dir> <template-type> <output-dir> <module-name> <framework-path>")
		fmt.Println("Example: verify-template ./cmd/gorp/cmd/templates multi-flat ./.tmp/verify-test test-project ..")
		os.Exit(1)
	}

	templatesDir := os.Args[1]
	templateType := os.Args[2]
	outputDir := os.Args[3]
	moduleName := os.Args[4]
	frameworkPath := os.Args[5]

	// 计算框架的绝对路径
	cwd, _ := os.Getwd()
	absFramework := filepath.Join(cwd, frameworkPath)
	absFramework, _ = filepath.Abs(absFramework)
	absOutput, _ := filepath.Abs(outputDir)

	// 从输出目录到框架目录的相对路径（用于 go.mod）
	relFramework, _ := filepath.Rel(absOutput, absFramework)
	relFramework = filepath.ToSlash(relFramework)

	// 对于 multi-independent，服务目录在 services/user 下，需要额外两级
	var serviceRelFramework string
	if templateType == "multi-independent" {
		// 从 services/user 目录到框架目录
		serviceDir := filepath.Join(absOutput, "services", "user")
		serviceRelFramework, _ = filepath.Rel(serviceDir, absFramework)
		serviceRelFramework = filepath.ToSlash(serviceRelFramework)
	} else {
		serviceRelFramework = relFramework
	}

	// 模板数据
	data := map[string]any{
		"ModuleName":       moduleName,
		"ProjectName":      strings.ReplaceAll(moduleName, "/", "-"),
		"FrameworkModule":  "github.com/ngq/gorp",
		"FrameworkPath":    serviceRelFramework,
		"WithDB":           true,
		"WithSwagger":      true,
		"WithAuth":         true,
		"WithRBAC":         true,
		"WithAdmin":        true,
		"Backend":          "gorm",
		"HasDDD":           false,
		"HasGRPC":          false,
	}

	// 模板根目录
	templateRoot := filepath.Join(templatesDir, templateType, "project")

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	// 渲染模板
	if err := renderTemplateDir(os.DirFS(templateRoot), ".", outputDir, data); err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Template %s rendered to %s\n", templateType, outputDir)
	fmt.Printf("Framework path: %s\n", serviceRelFramework)
}

func renderTemplateDir(src fs.FS, srcRoot string, dstRoot string, data map[string]any) error {
	return fs.WalkDir(src, srcRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, srcRoot)
		rel = strings.TrimPrefix(rel, "/")
		rel = strings.TrimPrefix(rel, "\\")

		if d.IsDir() {
			if rel == "" {
				return os.MkdirAll(dstRoot, 0755)
			}
			return os.MkdirAll(filepath.Join(dstRoot, rel), 0755)
		}

		outPath := filepath.Join(dstRoot, rel)
		if strings.HasSuffix(outPath, ".tmpl") {
			outPath = strings.TrimSuffix(outPath, ".tmpl")
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		b, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".tmpl") {
			tpl, err := template.New(rel).Parse(string(b))
			if err != nil {
				return fmt.Errorf("parse template %s: %w", rel, err)
			}
			var buf bytes.Buffer
			if err := tpl.Execute(&buf, data); err != nil {
				return fmt.Errorf("execute template %s: %w", rel, err)
			}
			b = buf.Bytes()
		}

		if len(bytes.TrimSpace(b)) == 0 {
			return nil
		}

		return os.WriteFile(outPath, b, 0644)
	})
}