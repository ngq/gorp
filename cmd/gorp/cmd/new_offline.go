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

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/spf13/cobra"
)

//go:embed all:templates/project all:templates/golayout all:templates/multi-flat-wire all:templates/multi-independent
var projectTemplateFS embed.FS

var newOfflineTemplate string
var newOfflineBackend string
var newOfflineWithDB bool
var newOfflineWithSwagger bool
var newOfflineFrameworkVersion string
var newOfflineHTTP string
var newOfflineLocalDev bool

const defaultFrameworkVersion = "v0.1.0"

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
	// - 默认不生成 replace，go.mod 直接引用 GitHub 发布的框架模块；
	// - 如果用户指定了 --framework-version，使用指定版本；
	// - 如果用户指定了 --local，进入本地开发模式，使用 replace 指向本地框架源码路径。
	needFrameworkPath := newOfflineLocalDev
	project, name, err := promptProjectInput(in, cmd.OutOrStdout(), frameworkDefault, needFrameworkPath)
	if err != nil {
		return err
	}

	// 确定框架版本：优先使用 --framework-version，否则用默认版本
	fwVersion := newOfflineFrameworkVersion
	if fwVersion == "" {
		fwVersion = defaultFrameworkVersion
	}
	project.FrameworkVersion = fwVersion
	// 非本地开发模式下清空 path，确保不生成 replace
	if !newOfflineLocalDev {
		project.FrameworkPath = ""
	}
	if newOfflineBackend == "" {
		newOfflineBackend = string(datacontract.RuntimeBackendGorm)
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

	// HTTP 模式解析：--http 参数，不传则使用默认值 contract
	// 治理模式由模板类型决定：golayout → mono，multi-* → micro
	//
	// HTTP mode: --http flag, defaults to contract
	// Governance mode is determined by template: golayout → mono, multi-* → micro
	project.HTTPMode = normalizeHTTPMode(newOfflineHTTP)
	project.Governance = defaultGovernanceByTemplate(newOfflineTemplate)
	project.GovernanceMode = expandGovernanceMode(project.Governance, project.HTTPMode)

	folder, err := prepareScaffoldTargetDir(cwd, name)
	if err != nil {
		return err
	}

	data := buildScaffoldData(project)
	templateRoot := resolveOfflineTemplateRoot(newOfflineTemplate)
	if err := renderTemplateProject(projectTemplateFS, templateRoot, folder, data); err != nil {
		return err
	}

	printScaffoldNext(cmd.OutOrStdout(), folder, newOfflineTemplate)
	return nil
}

func init() {
	newCmd.RunE = runNewEmbedded
	newCmd.Args = cobra.MaximumNArgs(1)
	newCmd.ValidArgs = []string{newIntentMultiWire}
	newCmd.Flags().StringVar(&newOfflineTemplate, "template", starterTemplateGoLayout, "starter template: golayout, multi-flat-wire, multi-independent")
	newCmd.Flags().StringVar(&newOfflineBackend, "backend", string(datacontract.RuntimeBackendGorm), "starter backend: gorm|ent")
	newCmd.Flags().BoolVar(&newOfflineWithDB, "with-db", true, "include DB sample and CRUD example")
	newCmd.Flags().BoolVar(&newOfflineWithSwagger, "with-swagger", true, "enable swagger config in generated starter")
	newCmd.Flags().BoolVar(&newOfflineLocalDev, "local", false, "use local framework path with replace directive for development")
	newCmd.Flags().StringVar(&newOfflineFrameworkVersion, "framework-version", "", "framework version (e.g., v0.1.0), default "+defaultFrameworkVersion)
	newCmd.Flags().StringVar(&newOfflineHTTP, "http", "", "HTTP mode: contract (default), gin")
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
