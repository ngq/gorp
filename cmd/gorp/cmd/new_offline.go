package cmd

import (
	"bufio"
	"bytes"
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ngq/gorp/framework/contract"
	"github.com/spf13/cobra"
)

//go:embed all:templates/project all:templates/golayout all:templates/golayout-wire all:templates/multi-flat all:templates/multi-flat-wire
var projectTemplateFS embed.FS

var newOfflineTemplate string
var newOfflinePreset string
var newOfflineBackend string
var newOfflineWithDB bool
var newOfflineWithSwagger bool
var newOfflineWithAuth bool
var newOfflineWithRBAC bool
var newOfflineWithAdmin bool

var newOfflineCmd = &cobra.Command{
	Use:   "offline",
	Short: "Create a new project from embedded template (offline)",
	Long: `Create a new project from embedded starter templates without network access.

Template options:
  - base            : minimal skeleton for custom structure
  - golayout        : standard single-service template
  - golayout-wire   : single-service template with Wire assembly
  - multi-flat      : default multi-service template
  - multi-flat-wire : multi-service template with Wire assembly

Preset recommendation:
  - golayout-basic / golayout-enterprise: user-facing presets for the golayout template

Important:
  - When you specify --preset without --template, the template is automatically inferred.
  - For example: --preset golayout-basic automatically selects --template golayout.

If you are not sure which template to pick:
  - Single service: --template golayout or --template golayout-wire
  - Multi-service: --template multi-flat or --template multi-flat-wire`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("template") && newOfflinePreset != "" {
			if inferred := templateFromPreset(newOfflinePreset); inferred != "" {
				newOfflineTemplate = inferred
			}
		}
		if err := validateStarterTemplate(newOfflineTemplate); err != nil {
			return err
		}
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		in := bufio.NewReader(cmd.InOrStdin())
		frameworkDefault, err := repoRootFromCWD()
		if err != nil {
			frameworkDefault = cwd
		}
		project, name, err := promptProjectInput(in, cmd.OutOrStdout(), frameworkDefault, true)
		if err != nil {
			return err
		}
		if newOfflineBackend == "" {
			newOfflineBackend = string(contract.RuntimeBackendGorm)
		}
		project.Backend = newOfflineBackend

		presetProject := project
		if err := applyStarterPreset(newOfflinePreset, &presetProject); err != nil {
			return err
		}
		if cmd.Flags().Changed("with-db") {
			project.WithDB = newOfflineWithDB
		} else {
			project.WithDB = presetProject.WithDB
		}
		if cmd.Flags().Changed("with-swagger") {
			project.WithSwagger = newOfflineWithSwagger
		} else {
			project.WithSwagger = presetProject.WithSwagger
		}
		if cmd.Flags().Changed("with-auth") {
			project.WithAuth = newOfflineWithAuth
		} else {
			project.WithAuth = presetProject.WithAuth
		}
		if cmd.Flags().Changed("with-rbac") {
			project.WithRBAC = newOfflineWithRBAC
		} else {
			project.WithRBAC = presetProject.WithRBAC
		}
		if cmd.Flags().Changed("with-admin") {
			project.WithAdmin = newOfflineWithAdmin
		} else {
			project.WithAdmin = presetProject.WithAdmin
		}

		folder, err := prepareScaffoldTargetDir(cwd, name)
		if err != nil {
			return err
		}

		data := buildScaffoldData(project)
		templateRoot := resolveOfflineTemplateRoot(newOfflineTemplate)
		if err := renderTemplateProject(projectTemplateFS, templateRoot, folder, data); err != nil {
			return err
		}

		printScaffoldNext(cmd.OutOrStdout(), folder)
		return nil
	},
}

func init() {
	newOfflineCmd.Flags().StringVar(&newOfflineTemplate, "template", starterTemplateBase, "starter template: base, golayout, golayout-wire, multi-flat, multi-flat-wire")
	newOfflineCmd.Flags().StringVar(&newOfflinePreset, "preset", "", "user preset: golayout-basic or golayout-enterprise")
	newOfflineCmd.Flags().StringVar(&newOfflineBackend, "backend", string(contract.RuntimeBackendGorm), "starter backend: gorm|ent")
	newOfflineCmd.Flags().BoolVar(&newOfflineWithDB, "with-db", true, "include DB sample and CRUD example")
	newOfflineCmd.Flags().BoolVar(&newOfflineWithSwagger, "with-swagger", true, "enable swagger config in generated starter")
	newOfflineCmd.Flags().BoolVar(&newOfflineWithAuth, "with-auth", true, "include auth/session starter")
	newOfflineCmd.Flags().BoolVar(&newOfflineWithRBAC, "with-rbac", true, "include Casbin RBAC starter")
	newOfflineCmd.Flags().BoolVar(&newOfflineWithAdmin, "with-admin", true, "include protected admin routes and RBAC management APIs")
	newCmd.AddCommand(newOfflineCmd)
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
				return os.MkdirAll(dstRoot, 0o755)
			}
			return os.MkdirAll(filepath.Join(dstRoot, rel), 0o755)
		}

		outPath := filepath.Join(dstRoot, rel)
		if strings.HasSuffix(outPath, ".tmpl") {
			outPath = strings.TrimSuffix(outPath, ".tmpl")
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}

		b, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".tmpl") {
			tpl, err := template.New(rel).Parse(string(b))
			if err != nil {
				return err
			}
			var buf bytes.Buffer
			if err := tpl.Execute(&buf, data); err != nil {
				return err
			}
			b = buf.Bytes()
		}

		// 中文说明：
		// - 如果渲染后的内容为空，跳过写入；
		// - 这样可以避免生成空文件（例如 ent 相关文件在 gorm backend 时不需要生成）。
		if len(bytes.TrimSpace(b)) == 0 {
			return nil
		}

		return os.WriteFile(outPath, b, 0o644)
	})
}
