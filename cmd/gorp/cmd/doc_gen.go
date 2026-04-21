package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	docOut         string
	docEnv         string
	docStdout      bool
	docCheck       bool
	docProjectRoot string
	docNoRedact    bool
)

var docGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate operator manual markdown under docs/manual/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root := docProjectRoot
		if root == "" {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			root = wd
		}

		outBase := docOut
		if outBase == "" {
			outBase = filepath.Join(root, "docs", "manual")
		}

		// 中文说明：
		// - `doc gen` 的核心是把“CLI 结构 + 配置结构”导出成静态 markdown。
		// - `--stdout` 适合预览，`--check` 适合 CI 校验，默认行为是写入 docs/manual/。
		// - 生成内容尽量保持确定性，方便后续做 diff 审查。
		files, err := generateManualFiles(root, docEnv, !docNoRedact)
		if err != nil {
			return err
		}

		if docStdout {
			// Print index.md by default, plus referenced pages for convenience.
			order := []string{"index.md", "cli.md", "config.md"}
			for i, name := range order {
				if i > 0 {
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintln(cmd.OutOrStdout(), "---")
					fmt.Fprintln(cmd.OutOrStdout())
				}
				fmt.Fprintf(cmd.OutOrStdout(), "<!-- %s -->\n\n", name)
				_, err := cmd.OutOrStdout().Write(files[name])
				if err != nil {
					return err
				}
			}
			return nil
		}

		if docCheck {
			return checkManualFiles(outBase, files)
		}

		return writeManualFiles(outBase, files)
	},
}

func init() {
	docCmd.AddCommand(docGenCmd)
	docGenCmd.Flags().StringVar(&docOut, "out", "", "output directory (default <projectRoot>/docs/manual)")
	docGenCmd.Flags().StringVar(&docEnv, "env", "", "env name (if empty, generate base + all overlays matrix)")
	docGenCmd.Flags().BoolVar(&docStdout, "stdout", false, "write index.md to stdout only")
	docGenCmd.Flags().BoolVar(&docCheck, "check", false, "check generated content matches existing files (no write)")
	docGenCmd.Flags().StringVar(&docProjectRoot, "project-root", "", "project root path (default current working directory)")
	docGenCmd.Flags().BoolVar(&docNoRedact, "no-redact", false, "do not redact secret-like config values")
}

// ---- generation ----

type commandDoc struct {
	Path     string
	Short    string
	Flags    []flagDoc
	InhFlags []flagDoc
	Children []commandDoc
}

type flagDoc struct {
	Name      string
	Shorthand string
	Type      string
	Default   string
	Usage     string
}

func generateManualFiles(projectRoot, env string, redact bool) (map[string][]byte, error) {
	// 中文说明：
	// - 这里统一组装手册的三个核心页面：index / cli / config。
	// - 之所以先在内存里全部生成，再统一写文件，是为了便于 stdout/check/write 三种模式复用同一份结果。
	cliMd, err := renderCLIDoc(rootCmd)
	if err != nil {
		return nil, err
	}
	configMd, err := renderConfigDoc(projectRoot, env, redact)
	if err != nil {
		return nil, err
	}
	indexMd := renderIndexDoc()

	files := map[string][]byte{
		"index.md":  []byte(indexMd),
		"cli.md":    []byte(cliMd),
		"config.md": []byte(configMd),
	}

	// normalize newlines
	for k, v := range files {
		files[k] = []byte(strings.ReplaceAll(string(v), "\r\n", "\n"))
	}
	return files, nil
}

func renderIndexDoc() string {
	// Use a normal quoted string so we can include markdown backticks.
	return "# 操作手册（Operator Manual）\n\n" +
		"本手册由 `gorp doc gen` 自动生成（并可在此基础上补充说明）。\n\n" +
		"> 如果你是第一次接触 gorp，建议先阅读：[用户上手路径：安装 gorp 并创建自己的项目](onboarding.md)。\n\n" +
		"## Quick start（快速开始）\n\n" +
		"### 默认使用路径\n\n" +
		"- 安装 CLI：`go install github.com/ngq/gorp/cmd/gorp@latest`\n" +
		"- 生成项目：`gorp new`\n" +
		"- 进入生成项目后开发与启动：`go run ./cmd/app`\n" +
		"- 验证：`GET /healthz`\n\n" +
		"### 常见工具链入口\n\n" +
		"- 创建项目：`go run ./cmd/gorp new --help`\n" +
		"- 模板治理：`go run ./cmd/gorp template version`\n" +
		"- 协议/模型生成：`go run ./cmd/gorp proto --help`、`go run ./cmd/gorp model --help`\n" +
		"- 当前 CLI 主线已不再把 `app / grpc / cron / build / dev / deploy` 作为 starter 项目的公开 runtime 路径；服务启动应通过生成项目自己的 `cmd/*/main.go` 入口\n\n" +
		"### 启动项目\n\n" +
		"- 建议在生成项目内通过自己的 `cmd/*/main.go` 入口启动，例如：`go run ./cmd/app`\n" +
		"- 验证：`GET /healthz`\n\n" +
		"### 生成 API 文档\n\n" +
		"- Swagger2: `go run ./cmd/gorp swagger gen`\n" +
		"- OpenAPI3: `go run ./cmd/gorp openapi gen`\n" +
		"- 打开 Swagger UI：`GET /swagger/index.html`\n\n" +
		"### 基于数据库生成代码\n\n" +
		"- DB 连通性：`go run ./cmd/gorp model test`\n" +
		"- 生成模型：`go run ./cmd/gorp model gen --table users`\n" +
		"- 生成 CRUD 骨架：`go run ./cmd/gorp model api --table users`\n\n" +
		"### 生成本手册\n\n" +
		"- `go run ./cmd/gorp doc gen`（输出到 `docs/manual/`）\n\n" +
		"## 目录\n\n" +
		"- [CLI 命令参考](cli.md)\n" +
		"- [配置参考](config.md)\n" +
		"- [开发手册（手工维护）](dev.md)\n" +
		"- [框架作者手册（手工维护）](author.md)\n"
}

func renderCLIDoc(root *cobra.Command) (string, error) {
	// Ensure default help command/flag exist deterministically.
	root.InitDefaultHelpCmd()
	root.InitDefaultHelpFlag()

	var b strings.Builder
	b.WriteString("# CLI 命令参考\n\n")
	b.WriteString("> 说明：本页主要由 `gorp doc gen` 自动生成，用于作为命令索引与参数查阅页。\n")
	b.WriteString(">\n")
	b.WriteString("> 阅读建议：\n")
	b.WriteString(">\n")
	b.WriteString("> - 如果你是第一次使用本项目，建议先看：\n")
	b.WriteString(">   - [用户上手路径：安装 gorp 并创建自己的项目](onboarding.md)\n")
	b.WriteString(">   - [教程：利用脚手架快速开展业务开发](tutorial-scaffold-quickstart.md)\n")
	b.WriteString("> - 如果你已经知道自己要用哪个命令，再回到本页查参数最合适。\n")
	b.WriteString(">\n")
	b.WriteString("> 约定：\n")
	b.WriteString(">\n")
	b.WriteString("> - `gorp new` / `template` / `proto` / `model`：项目创建、模板治理与代码生成主入口\n")
	b.WriteString("> - `gorp doc` / `swagger` / `openapi`：文档生成相关\n")
	b.WriteString("> - 当前 CLI 主线已不再把 `app / grpc / cron / build / dev / deploy` 作为 starter 项目的公开 runtime 路径；服务启动应通过生成项目自己的 `cmd/*/main.go` 入口\n\n")

	docs := collectCommands(root)
	for _, d := range docs {
		renderCommand(&b, d)
	}
	return b.String(), nil
}

func collectCommands(root *cobra.Command) []commandDoc {
	out := make([]commandDoc, 0, 32)
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd == nil {
			return
		}
		// Skip hidden commands except root.
		if cmd != root && cmd.Hidden {
			return
		}
		out = append(out, commandToDoc(cmd))

		children := cmd.Commands()
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name() < children[j].Name()
		})
		for _, c := range children {
			walk(c)
		}
	}
	walk(root)
	return out
}

func commandToDoc(cmd *cobra.Command) commandDoc {
	return commandDoc{
		Path:     cmd.CommandPath(),
		Short:    cmd.Short,
		Flags:    collectFlags(cmd.LocalFlags()),
		InhFlags: collectFlags(cmd.InheritedFlags()),
	}
}

func collectFlags(fs *pflag.FlagSet) []flagDoc {
	flags := make([]flagDoc, 0)
	if fs == nil {
		return flags
	}
	fs.VisitAll(func(f *pflag.Flag) {
		// Skip the ubiquitous help flag in per-command sections.
		if f.Name == "help" {
			return
		}
		flags = append(flags, flagDoc{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		})
	})
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
	return flags
}

func renderCommand(b *strings.Builder, d commandDoc) {
	b.WriteString("## ")
	b.WriteString(d.Path)
	b.WriteString("\n\n")
	if d.Short != "" {
		b.WriteString(d.Short)
		b.WriteString("\n\n")
	}
	if len(d.Flags) > 0 {
		b.WriteString("**Flags**\n\n")
		for _, f := range d.Flags {
			b.WriteString("- `")
			if f.Shorthand != "" {
				b.WriteString("-")
				b.WriteString(f.Shorthand)
				b.WriteString(", ")
			}
			b.WriteString("--")
			b.WriteString(f.Name)
			b.WriteString(" (")
			b.WriteString(f.Type)
			if f.Default != "" {
				b.WriteString(", default: ")
				b.WriteString(f.Default)
			}
			b.WriteString(")` ")
			b.WriteString(" ")
			b.WriteString(strings.TrimSpace(f.Usage))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
}

func renderConfigDoc(projectRoot, env string, redact bool) (string, error) {
	files, order, err := discoverConfigFiles(projectRoot, env)
	if err != nil {
		return "", err
	}

	cfg, err := mergedConfigForDocs(projectRoot, env)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("# 配置参考\n\n")
	b.WriteString("> 说明：本页由 `gorp doc gen` 自动生成，用于帮助你快速查看当前工程可见配置结构。\n\n")

	if len(order) > 0 {
		b.WriteString("## 加载到的配置文件\n\n")
		for _, f := range order {
			if rel, err := filepath.Rel(projectRoot, f); err == nil {
				b.WriteString("- `")
				b.WriteString(filepath.ToSlash(rel))
				b.WriteString("`\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("## 合并后配置（YAML）\n\n")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	text := string(data)
	if redact {
		text = redactYAMLText(text)
	}
	b.WriteString("```yaml\n")
	b.WriteString(text)
	if !strings.HasSuffix(text, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")

	b.WriteString("## 原始配置文件片段\n\n")
	for _, f := range order {
		raw := files[f]
		if redact {
			raw = []byte(redactYAMLText(string(raw)))
		}
		rel := f
		if r, err := filepath.Rel(projectRoot, f); err == nil {
			rel = filepath.ToSlash(r)
		}
		b.WriteString("### `")
		b.WriteString(rel)
		b.WriteString("`\n\n")
		b.WriteString("```yaml\n")
		b.Write(raw)
		if len(raw) == 0 || raw[len(raw)-1] != '\n' {
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}

	return b.String(), nil
}

func discoverConfigFiles(projectRoot, env string) (map[string][]byte, []string, error) {
	files := map[string][]byte{}
	order := []string{}

	configDir := filepath.Join(projectRoot, "config")
	entries, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return files, order, nil
		}
		return nil, nil, err
	}

	// base files: app.yaml first, then other *.yaml excluding app.<env>.yaml
	base := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}
		if matched, _ := regexp.MatchString(`^app\.[^.]+\.yaml$`, name); matched {
			continue
		}
		base = append(base, name)
	}
	sort.Slice(base, func(i, j int) bool {
		if base[i] == "app.yaml" {
			return true
		}
		if base[j] == "app.yaml" {
			return false
		}
		return base[i] < base[j]
	})
	for _, name := range base {
		path := filepath.Join(configDir, name)
		bs, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		files[path] = bs
		order = append(order, path)
	}

	if env != "" {
		overlay := filepath.Join(configDir, fmt.Sprintf("app.%s.yaml", env))
		if bs, err := os.ReadFile(overlay); err == nil {
			files[overlay] = bs
			order = append(order, overlay)
		}
		envDir := filepath.Join(configDir, env)
		if des, err := os.ReadDir(envDir); err == nil {
			names := make([]string, 0)
			for _, de := range des {
				if de.IsDir() {
					continue
				}
				if strings.HasSuffix(de.Name(), ".yaml") {
					names = append(names, de.Name())
				}
			}
			sort.Strings(names)
			for _, name := range names {
				path := filepath.Join(envDir, name)
				bs, err := os.ReadFile(path)
				if err != nil {
					return nil, nil, err
				}
				files[path] = bs
				order = append(order, path)
			}
		}
	}
	return files, order, nil
}

func mergedConfigForDocs(projectRoot, env string) (map[string]any, error) {
	files, order, err := discoverConfigFiles(projectRoot, env)
	if err != nil {
		return nil, err
	}
	merged := map[string]any{}
	for _, f := range order {
		var m map[string]any
		if err := yaml.Unmarshal(files[f], &m); err != nil {
			return nil, fmt.Errorf("parse %s: %w", f, err)
		}
		merged = mergeMaps(merged, m)
	}
	return merged, nil
}

func mergeMaps(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}
	for k, v := range src {
		if vm, ok := v.(map[string]any); ok {
			if dv, ok := dst[k].(map[string]any); ok {
				dst[k] = mergeMaps(dv, vm)
			} else {
				dst[k] = mergeMaps(map[string]any{}, vm)
			}
			continue
		}
		dst[k] = v
	}
	return dst
}

var redactKeyPattern = regexp.MustCompile(`(?i)(password|secret|token|key|dsn|uri)$`)
var redactValuePattern = regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z\-_]{35}|(?:(?:eyJ)[A-Za-z0-9_\-.]+)|(?:-----BEGIN [A-Z ]+-----))`)

func redactYAMLText(s string) string {
	// redact values for suspicious keys in YAML lines: key: value
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			if redactValuePattern.MatchString(line) {
				lines[i] = redactValuePattern.ReplaceAllString(line, "***REDACTED***")
			}
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if redactKeyPattern.MatchString(key) && val != "" {
			lines[i] = line[:idx+1] + " ***REDACTED***"
			continue
		}
		if redactValuePattern.MatchString(val) {
			lines[i] = line[:idx+1] + " ***REDACTED***"
		}
	}
	return strings.Join(lines, "\n")
}

// ---- check / write helpers ----

func writeManualFiles(outBase string, files map[string][]byte) error {
	if err := os.MkdirAll(outBase, 0o755); err != nil {
		return err
	}
	for name, content := range files {
		path := filepath.Join(outBase, name)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func checkManualFiles(outBase string, files map[string][]byte) error {
	mismatches := make([]string, 0)
	for name, want := range files {
		path := filepath.Join(outBase, name)
		have, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				mismatches = append(mismatches, fmt.Sprintf("missing %s", name))
				continue
			}
			return err
		}
		if !bytes.Equal(normalizeNL(have), normalizeNL(want)) {
			mismatches = append(mismatches, fmt.Sprintf("outdated %s", name))
		}
	}
	if len(mismatches) > 0 {
		return fmt.Errorf("manual docs are not up to date: %s", strings.Join(mismatches, ", "))
	}
	return nil
}

func normalizeNL(b []byte) []byte { return []byte(strings.ReplaceAll(string(b), "\r\n", "\n")) }

// ---- small utilities ----

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
