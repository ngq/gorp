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

//go:embed all:templates/project all:templates/golayout all:templates/golayout-wire all:templates/multi-flat all:templates/multi-flat-wire all:templates/multi-independent
var projectTemplateFS embed.FS

var newOfflineTemplate string
var newOfflineBackend string
var newOfflineWithDB bool
var newOfflineWithSwagger bool
var newOfflineFrameworkVersion string

func runNewEmbedded(cmd *cobra.Command, args []string) error {
	intent, err := parseNewIntent(args)
	if err != nil {
		return err
	}
	intentTemplate, intentStarter := resolveOfflineIntentDefaults(intent)

	// 中文说明：
	// - 裸 `gorp new` 现在默认走单服务 basic 起步路径，而不是继续落到 base；
	// - 位置参数只表达高频模板结构意图，例如 wire / multi / multi-wire；
	// - 显式 `--template` 仍然拥有更高优先级，但 starter 不再公开承诺 auth/rbac/admin 业务能力。
	if !cmd.Flags().Changed("template") {
		newOfflineTemplate = intentTemplate
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

	// 中文说明：
	// - 如果用户指定了 --framework-version，则使用 GitHub 版本，不需要 replace；
	// - 否则使用本地 replace 模式，询问框架源码路径。
	needFrameworkPath := newOfflineFrameworkVersion == ""
	project, name, err := promptProjectInput(in, cmd.OutOrStdout(), frameworkDefault, needFrameworkPath)
	if err != nil {
		return err
	}

	// 如果指定了版本，覆盖默认的版本和路径
	if newOfflineFrameworkVersion != "" {
		project.FrameworkVersion = newOfflineFrameworkVersion
		project.FrameworkPath = "" // 清空路径，不生成 replace
	}
	if newOfflineBackend == "" {
		newOfflineBackend = string(contract.RuntimeBackendGorm)
	}
	project.Backend = newOfflineBackend

	starterProject := project
	applyIntentDefaults(intentStarter, &starterProject)
	if cmd.Flags().Changed("with-db") {
		project.WithDB = newOfflineWithDB
	} else {
		project.WithDB = starterProject.WithDB
	}
	if cmd.Flags().Changed("with-swagger") {
		project.WithSwagger = newOfflineWithSwagger
	} else {
		project.WithSwagger = starterProject.WithSwagger
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
}

func init() {
	newCmd.RunE = runNewEmbedded
	newCmd.Args = cobra.MaximumNArgs(1)
	newCmd.ValidArgs = []string{newIntentWire, newIntentMulti, newIntentMultiWire}
	newCmd.Flags().StringVar(&newOfflineTemplate, "template", starterTemplateGoLayout, "starter template: base, golayout, golayout-wire, multi-flat, multi-flat-wire")
	newCmd.Flags().StringVar(&newOfflineBackend, "backend", string(contract.RuntimeBackendGorm), "starter backend: gorm|ent")
	newCmd.Flags().BoolVar(&newOfflineWithDB, "with-db", true, "include DB sample and CRUD example")
	newCmd.Flags().BoolVar(&newOfflineWithSwagger, "with-swagger", true, "enable swagger config in generated starter")
	newCmd.Flags().StringVar(&newOfflineFrameworkVersion, "framework-version", "", "framework version (e.g., v0.1.0), if set, no replace directive will be generated")
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
