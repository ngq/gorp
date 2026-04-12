package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// templateDiffCmd 比较当前项目与模板的差异。
//
// 中文说明：
// - 用于检查项目与最新模板之间的差异；
// - 帮助用户决定是否需要升级模板；
// - 输出差异文件列表，不自动修改任何文件。
var templateDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare project with latest template",
	Long: `Compare current project structure with the latest template.

This command helps you:
  1. See which files are different from the template
  2. Decide whether to upgrade your project
  3. Preview changes before applying them

Output:
  - Files only in project (custom files)
  - Files only in template (new features)
  - Files different between project and template

This is a read-only operation. No files are modified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查是否在项目目录中
		if !fileExists("go.mod") {
			return fmt.Errorf("not in a project directory (go.mod not found)")
		}

		// 检测当前项目使用的模板类型
		templateType := detectProjectTemplateType()
		if templateType == "" {
			return fmt.Errorf("cannot detect project template type")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Detected template: %s\n\n", templateType)

		onlyProject, onlyTemplate, different, err := computeTemplateDiff(".", templateType)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Template Diff Report")
		fmt.Fprintln(cmd.OutOrStdout(), "===================")
		fmt.Fprintln(cmd.OutOrStdout())

		fmt.Fprintf(cmd.OutOrStdout(), "Only in project (%d):\n", len(onlyProject))
		for _, p := range onlyProject {
			fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", p)
		}
		fmt.Fprintln(cmd.OutOrStdout())

		fmt.Fprintf(cmd.OutOrStdout(), "Only in template (%d):\n", len(onlyTemplate))
		for _, p := range onlyTemplate {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", p)
		}
		fmt.Fprintln(cmd.OutOrStdout())

		fmt.Fprintf(cmd.OutOrStdout(), "Different content (%d):\n", len(different))
		for _, p := range different {
			fmt.Fprintf(cmd.OutOrStdout(), "  ! %s\n", p)
		}
		fmt.Fprintln(cmd.OutOrStdout())

		fmt.Fprintln(cmd.OutOrStdout(), "Next: run 'gorp template upgrade --dry-run' to preview an upgrade workflow.")
		return nil
	},
}

// templateUpgradeCmd 升级项目模板。
//
// 中文说明：
// - 用于将项目升级到最新模板版本；
// - 支持 --dry-run 预览变更；
// - 支持 --files 指定要升级的文件；
// - 默认不会覆盖用户修改过的文件（除非使用 --force）。
var templateUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade project to latest template",
	Long: `Upgrade project to the latest template version.

This command helps you:
  1. Update Dockerfile, Makefile, CI configs
  2. Add new template features to existing projects
  3. Keep project structure in sync with best practices

Safety features:
  - --dry-run: preview changes without modifying files
  - --files: specify which files to upgrade
  - --force: overwrite files even if they were modified

Recommended workflow:
  1. Run 'gorp template diff' to see what's different
  2. Run 'gorp template upgrade --dry-run' to preview changes
  3. Run 'gorp template upgrade' to apply changes
  4. Review and test the updated project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")
		files, _ := cmd.Flags().GetStringSlice("files")

		if !fileExists("go.mod") {
			return fmt.Errorf("not in a project directory (go.mod not found)")
		}

		templateType := detectProjectTemplateType()
		if templateType == "" {
			return fmt.Errorf("cannot detect project template type")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Detected template: %s\n", templateType)
		fmt.Fprintf(cmd.OutOrStdout(), "Dry run: %v\n", dryRun)
		fmt.Fprintf(cmd.OutOrStdout(), "Force: %v\n", force)
		if len(files) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Files: %v\n", files)
		}
		fmt.Fprintln(cmd.OutOrStdout())

		if dryRun {
			onlyProject, onlyTemplate, different, err := computeTemplateDiff(".", templateType)
			if err != nil {
				return err
			}
			selectedOnlyTemplate := filterTemplateDiffFiles(onlyTemplate, files)
			selectedDifferent := filterTemplateDiffFiles(different, files)
			fmt.Fprintln(cmd.OutOrStdout(), "Dry-run mode: previewing upgrade workflow from real diff results")
			fmt.Fprintln(cmd.OutOrStdout(), "====================================================")
			fmt.Fprintln(cmd.OutOrStdout())
			if len(files) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Scoped by --files: %v\n\n", files)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Files only in template (%d):\n", len(selectedOnlyTemplate))
			for _, p := range selectedOnlyTemplate {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", p)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintf(cmd.OutOrStdout(), "Files with different content (%d):\n", len(selectedDifferent))
			for _, p := range selectedDifferent {
				fmt.Fprintf(cmd.OutOrStdout(), "  ! %s\n", p)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			if len(selectedOnlyTemplate) == 0 && len(selectedDifferent) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No upgrade candidates found from current template diff.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Recommended next steps:")
				fmt.Fprintln(cmd.OutOrStdout(), "  1. Review the files above")
				fmt.Fprintln(cmd.OutOrStdout(), "  2. Commit or stash your current changes")
				fmt.Fprintln(cmd.OutOrStdout(), "  3. Re-run with --files to narrow the target set if needed")
			}
			if len(files) == 0 && len(onlyProject) > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "Note: files only in project are treated as custom files and are not upgrade targets by default.")
			}
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Template upgrade is a destructive operation.")
			fmt.Fprintln(cmd.OutOrStdout(), "Please use --dry-run first to preview changes.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "To proceed with upgrade:")
			fmt.Fprintln(cmd.OutOrStdout(), "  1. Commit or stash your current changes")
			fmt.Fprintln(cmd.OutOrStdout(), "  2. Run: gorp template upgrade --dry-run")
			fmt.Fprintln(cmd.OutOrStdout(), "  3. Review the changes")
			fmt.Fprintln(cmd.OutOrStdout(), "  4. Run: gorp template upgrade --force")
		}

		return nil
	},
}

var templateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show template version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Template Version Information")
		fmt.Fprintln(cmd.OutOrStdout(), "============================")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Current CLI embeds the following templates:")
		fmt.Fprintln(cmd.OutOrStdout(), "  - base: minimal skeleton")
		fmt.Fprintln(cmd.OutOrStdout(), "  - golayout: default lightweight-DDD-first single-service template")
		fmt.Fprintln(cmd.OutOrStdout(), "  - multi-flat: default multi-service template")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Templates are embedded at build time.")
		fmt.Fprintln(cmd.OutOrStdout(), "Upgrade CLI to get latest templates.")
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateDiffCmd)
	templateCmd.AddCommand(templateUpgradeCmd)
	templateCmd.AddCommand(templateVersionCmd)

	templateUpgradeCmd.Flags().Bool("dry-run", false, "preview changes without modifying files")
	templateUpgradeCmd.Flags().Bool("force", false, "overwrite files even if they were modified")
	templateUpgradeCmd.Flags().StringSlice("files", []string{}, "specific files to upgrade (comma-separated)")
}

func listEmbedTemplateFiles(src fs.FS, root string) ([]string, error) {
	var files []string
	err := fs.WalkDir(src, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasSuffix(rel, ".tmpl") {
			rel = strings.TrimSuffix(rel, ".tmpl")
		}
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func listProjectFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := filepath.Base(path)
			if name == ".git" || name == ".tmp" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func diffFileSets(projectFiles, templateFiles []string) (onlyProject, onlyTemplate, common []string) {
	pmap := make(map[string]struct{}, len(projectFiles))
	tmap := make(map[string]struct{}, len(templateFiles))
	for _, p := range projectFiles {
		pmap[p] = struct{}{}
	}
	for _, t := range templateFiles {
		tmap[t] = struct{}{}
	}
	for _, p := range projectFiles {
		if _, ok := tmap[p]; ok {
			common = append(common, p)
		} else {
			onlyProject = append(onlyProject, p)
		}
	}
	for _, t := range templateFiles {
		if _, ok := pmap[t]; !ok {
			onlyTemplate = append(onlyTemplate, t)
		}
	}
	sort.Strings(onlyProject)
	sort.Strings(onlyTemplate)
	sort.Strings(common)
	return
}

func resolveEmbeddedTemplatePath(src fs.FS, root, rel string) (string, error) {
	rel = filepath.ToSlash(rel)
	p := filepath.ToSlash(filepath.Join(root, rel))
	if _, err := fs.Stat(src, p); err == nil {
		return p, nil
	}
	pTmpl := p + ".tmpl"
	if _, err := fs.Stat(src, pTmpl); err == nil {
		return pTmpl, nil
	}
	return "", fmt.Errorf("template file not found: %s", rel)
}

func filterTemplateDiffFiles(items, files []string) []string {
	if len(files) == 0 {
		return items
	}
	wanted := make(map[string]struct{}, len(files))
	for _, f := range files {
		wanted[filepath.ToSlash(strings.TrimSpace(f))] = struct{}{}
	}
	out := make([]string, 0)
	for _, item := range items {
		if _, ok := wanted[filepath.ToSlash(item)]; ok {
			out = append(out, item)
		}
	}
	return out
}

func sameFileContent(src fs.FS, templatePath, projectPath string, data map[string]any) (bool, error) {
	tplBytes, err := fs.ReadFile(src, filepath.ToSlash(templatePath))
	if err != nil {
		return false, err
	}
	if strings.HasSuffix(filepath.ToSlash(templatePath), ".tmpl") {
		rendered, err := renderEmbeddedTemplateContent(filepath.Base(templatePath), tplBytes, data)
		if err != nil {
			return false, err
		}
		tplBytes = rendered
	}
	projBytes, err := os.ReadFile(projectPath)
	if err != nil {
		return false, err
	}
	return bytes.Equal(normalizeTemplateCompareContent(tplBytes), normalizeTemplateCompareContent(projBytes)), nil
}

func computeTemplateDiff(projectRoot, templateType string) (onlyProject, onlyTemplate, different []string, err error) {
	tplRoot := resolveOfflineTemplateRoot(templateType)
	tplFiles, err := listEmbedTemplateFiles(projectTemplateFS, tplRoot)
	if err != nil {
		return nil, nil, nil, err
	}
	projFiles, err := listProjectFiles(projectRoot)
	if err != nil {
		return nil, nil, nil, err
	}
	onlyProject, onlyTemplate, common := diffFileSets(projFiles, tplFiles)
	compareData := buildTemplateCompareData(projectRoot)
	for _, rel := range common {
		tplPath, err := resolveEmbeddedTemplatePath(projectTemplateFS, tplRoot, rel)
		if err != nil {
			different = append(different, rel)
			continue
		}
		projPath := filepath.Join(projectRoot, filepath.FromSlash(rel))
		if same, _ := sameFileContent(projectTemplateFS, tplPath, projPath, compareData); !same {
			different = append(different, rel)
		}
	}
	sort.Strings(different)
	return onlyProject, onlyTemplate, different, nil
}

func buildTemplateCompareData(projectRoot string) map[string]any {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		absRoot = projectRoot
	}
	name := filepath.Base(absRoot)
	module := detectGoModulePath(filepath.Join(projectRoot, "go.mod"))
	frameworkModule, frameworkPath, frameworkVersion := detectFrameworkReference(filepath.Join(projectRoot, "go.mod"))
	withDB, withSwagger, withAuth, withRBAC, withAdmin := detectProjectFeatureFlags(projectRoot)
	return map[string]any{
		"Name":             name,
		"Module":           module,
		"FrameworkModule":  frameworkModule,
		"FrameworkPath":    filepath.ToSlash(frameworkPath),
		"FrameworkVersion": frameworkVersion,
		"WithDB":           withDB,
		"WithSwagger":      withSwagger,
		"WithAuth":         withAuth,
		"WithRBAC":         withRBAC,
		"WithAdmin":        withAdmin,
	}
}

func renderEmbeddedTemplateContent(name string, src []byte, data map[string]any) ([]byte, error) {
	tpl, err := template.New(name).Parse(string(src))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func normalizeTemplateCompareContent(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

func detectGoModulePath(goModPath string) string {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func detectFrameworkReference(goModPath string) (module, replacePath, version string) {
	module = "github.com/ngq/gorp"
	version = "v0.0.0"
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return module, replacePath, version
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "github.com/ngq/gorp ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				version = parts[1]
			}
		}
		if strings.HasPrefix(trimmed, "replace github.com/ngq/gorp =>") {
			replacePath = strings.TrimSpace(strings.TrimPrefix(trimmed, "replace github.com/ngq/gorp =>"))
		}
	}
	return module, replacePath, version
}

func detectProjectFeatureFlags(projectRoot string) (withDB, withSwagger, withAuth, withRBAC, withAdmin bool) {
	withSwagger = detectSwaggerEnabled(filepath.Join(projectRoot, "config", "app.yaml"))
	templateType := detectProjectTemplateType()
	switch templateType {
	case starterTemplateGoLayout:
		withDB = dirExists(filepath.Join(projectRoot, "internal", "data"))
		withAuth = fileExists(filepath.Join(projectRoot, "internal", "service", "auth.go"))
		withRBAC = fileExists(filepath.Join(projectRoot, "internal", "biz", "rbac.go"))
		withAdmin = fileExists(filepath.Join(projectRoot, "internal", "service", "admin.go"))
	default:
		withDB = false
		withAuth = false
		withRBAC = false
		withAdmin = false
	}
	return withDB, withSwagger, withAuth, withRBAC, withAdmin
}

func detectSwaggerEnabled(configPath string) bool {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return true
	}
	lines := strings.Split(string(content), "\n")
	inSwagger := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(raw, "swagger:") || line == "swagger:" {
			inSwagger = true
			continue
		}
		if inSwagger {
			if len(raw) > 0 && raw[0] != ' ' && raw[0] != '\t' {
				break
			}
			if strings.HasPrefix(line, "enable:") {
				return strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "enable:")), "true")
			}
		}
	}
	return true
}

func detectProjectTemplateType() string {
	if dirExists("internal/biz") && dirExists("internal/data") && dirExists("internal/service") {
		return starterTemplateGoLayout
	}
	if dirExists("app/http") {
		return starterTemplateBase
	}
	return ""
}

func walkDir(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
