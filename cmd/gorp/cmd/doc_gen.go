package cmd

import (
	"bytes"
	"crypto/sha256"
	"errors"
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
	docOut        string
	docEnv        string
	docStdout     bool
	docCheck      bool
	docProjectRoot string
	docNoRedact   bool
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
	Path        string
	Short       string
	Flags       []flagDoc
	InhFlags    []flagDoc
	Children    []commandDoc
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
		"## Quick start（快速开始）\n\n" +
		"### 启动 HTTP 服务\n\n" +
		"- `go run ./cmd/gorp app start`\n" +
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
	b.WriteString("> - 如果你是第一次使用本项目，建议先看：\n")
	b.WriteString(">   - [教程：启动项目并完成本地验证](tutorial-start.md)\n")
	b.WriteString(">   - [开发手册（dev.md）](dev.md)\n")
	b.WriteString("> - 如果你已经知道自己要用哪个命令，再回到本页查参数最合适。\n")
	b.WriteString(">\n")
	b.WriteString("> 约定：\n")
	b.WriteString("> - `gorp app`：HTTP 服务管理\n")
	b.WriteString("> - `gorp grpc`：gRPC 服务管理\n")
	b.WriteString("> - `gorp cron`：定时任务 worker 管理\n")
	b.WriteString("> - `gorp model`：数据库模型 / CRUD 骨架生成\n")
	b.WriteString("> - `gorp doc` / `swagger` / `openapi`：文档生成相关\n")
	b.WriteString("> - `gorp build` / `dev` / `deploy`：构建、开发联调、部署相关\n\n")

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
			b.WriteString("- ")
			b.WriteString(formatFlag(f))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(d.InhFlags) > 0 {
		b.WriteString("**Inherited flags**\n\n")
		for _, f := range d.InhFlags {
			b.WriteString("- ")
			b.WriteString(formatFlag(f))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
}

func formatFlag(f flagDoc) string {
	var parts []string
	if f.Shorthand != "" {
		parts = append(parts, fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name))
	} else {
		parts = append(parts, "--"+f.Name)
	}
	parts = append(parts, fmt.Sprintf("(%s, default: %s)", f.Type, quoteIfNeeded(f.Default)))
	if f.Usage != "" {
		parts = append(parts, f.Usage)
	}
	return strings.Join(parts, " ")
}

func quoteIfNeeded(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, " \t\n\r") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

// ---- config doc ----

type yamlDoc map[string]any

type kv struct {
	Key string
	Val any
}

func renderConfigDoc(projectRoot, env string, redact bool) (string, error) {
	basePath := filepath.Join(projectRoot, "config", "app.yaml")
	base, err := readYAMLFile(basePath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", basePath, err)
	}
	overlays, err := discoverOverlays(filepath.Join(projectRoot, "config"))
	if err != nil {
		return "", err
	}

	// 中文说明：
	// - config 文档有两种输出模式：
	//   1) 指定 --env：只输出该环境的最终生效配置；
	//   2) 未指定 --env：输出 base + 所有环境 overlay 的矩阵视图。
	// - 这样既适合开发排查单环境配置，也适合运维横向对比不同环境差异。
	var b strings.Builder
	b.WriteString("# Configuration Reference\n\n")
	b.WriteString("Config loading rule:\n")
	b.WriteString("- Base: merge all `config/*.yaml` except `config/app.<env>.yaml` (app.yaml first if present)\n")
	b.WriteString("- Env overlay (optional): merge `config/app.<env>.yaml` if it exists\n")
	b.WriteString("- Env dir overlay (optional): merge all `config/<env>/*.yaml`\n")
	b.WriteString("- `env(KEY)` placeholders are substituted from OS env during load\n")
	b.WriteString("(Environment variables are ignored in this document to keep output deterministic.)\n\n")

	if env != "" {
		o, ok := overlays[env]
		if !ok {
			return "", fmt.Errorf("env overlay not found: %s", env)
		}
		eff := deepMerge(cloneMap(base), o)
		rows := flatten(eff, "")
		sort.Slice(rows, func(i, j int) bool { return rows[i].Key < rows[j].Key })

		b.WriteString("## Effective config: ")
		b.WriteString(env)
		b.WriteString("\n\n")
		b.WriteString("| Key | Value |\n|---|---|\n")
		for _, r := range rows {
			b.WriteString("| ")
			b.WriteString(r.Key)
			b.WriteString(" | ")
			b.WriteString(renderValue(r.Key, r.Val, redact))
			b.WriteString(" |\n")
		}
		return b.String(), nil
	}

	// matrix view (base + each env effective)
	envs := make([]string, 0, len(overlays))
	for k := range overlays {
		envs = append(envs, k)
	}
	sort.Strings(envs)

	// compute flattened base and effective
	baseRows := flatten(base, "")
	baseMap := make(map[string]any, len(baseRows))
	for _, r := range baseRows {
		baseMap[r.Key] = r.Val
	}
	effMaps := make(map[string]map[string]any, len(envs))
	allKeys := make(map[string]struct{})
	for k := range baseMap {
		allKeys[k] = struct{}{}
	}
	for _, e := range envs {
		eff := deepMerge(cloneMap(base), overlays[e])
		rows := flatten(eff, "")
		m := make(map[string]any, len(rows))
		for _, r := range rows {
			m[r.Key] = r.Val
			allKeys[r.Key] = struct{}{}
		}
		effMaps[e] = m
	}
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteString("## Base + env overlays\n\n")
	b.WriteString("| Key | base (app.yaml) |")
	for _, e := range envs {
		b.WriteString(" ")
		b.WriteString(e)
		b.WriteString(" (effective) |")
	}
	b.WriteString("\n|")
	b.WriteString(strings.Repeat("---|", 2+len(envs)))
	b.WriteString("\n")
	for _, k := range keys {
		b.WriteString("| ")
		b.WriteString(k)
		b.WriteString(" | ")
		b.WriteString(renderValue(k, baseMap[k], redact))
		b.WriteString(" |")
		for _, e := range envs {
			b.WriteString(" ")
			b.WriteString(renderValue(k, effMaps[e][k], redact))
			b.WriteString(" |")
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

func discoverOverlays(configDir string) (map[string]yamlDoc, error) {
	out := map[string]yamlDoc{}
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`^app\.(.+)\.yaml$`)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := re.FindStringSubmatch(e.Name())
		if len(m) != 2 {
			continue
		}
		env := m[1]
		p := filepath.Join(configDir, e.Name())
		d, err := readYAMLFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		out[env] = d
	}
	return out, nil
}

func readYAMLFile(path string) (yamlDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	m, ok := normalizeYAML(v).(map[string]any)
	if !ok {
		return yamlDoc{}, nil
	}
	return yamlDoc(m), nil
}

func normalizeYAML(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, v := range x {
			out[k] = normalizeYAML(v)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(x))
		for k, v := range x {
			ks, _ := k.(string)
			out[ks] = normalizeYAML(v)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i := range x {
			out[i] = normalizeYAML(x[i])
		}
		return out
	default:
		return x
	}
}

func deepMerge(dst, overlay yamlDoc) yamlDoc {
	for k, ov := range overlay {
		if dv, ok := dst[k]; ok {
			dm, dok := dv.(map[string]any)
			om, ook := ov.(map[string]any)
			if dok && ook {
				merged := deepMerge(yamlDoc(dm), yamlDoc(om))
				dst[k] = map[string]any(merged)
				continue
			}
		}
		// scalar/array or missing: override
		dst[k] = ov
	}
	return dst
}

func cloneMap(in yamlDoc) yamlDoc {
	out := make(yamlDoc, len(in))
	for k, v := range in {
		out[k] = cloneValue(v)
	}
	return out
}

func cloneValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, v := range x {
			m[k] = cloneValue(v)
		}
		return m
	case []any:
		a := make([]any, len(x))
		for i := range x {
			a[i] = cloneValue(x[i])
		}
		return a
	default:
		return x
	}
}

func flatten(m yamlDoc, prefix string) []kv {
	out := make([]kv, 0)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch x := v.(type) {
		case map[string]any:
			out = append(out, flatten(yamlDoc(x), key)...)
		case []any:
			// arrays: store as a whole to keep it simple/deterministic
			out = append(out, kv{Key: key, Val: x})
		default:
			out = append(out, kv{Key: key, Val: x})
		}
	}
	return out
}

var secretKeyRe = regexp.MustCompile(`(?i)(password|secret|token|dsn)`) // case-insensitive

func renderValue(key string, v any, redact bool) string {
	if v == nil {
		return ""
	}
	if redact && secretKeyRe.MatchString(key) {
		return "******"
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return "\"\""
		}
		return fmt.Sprintf("%q", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int, int64, float64, uint64, uint:
		return fmt.Sprintf("%v", x)
	case []any:
		// keep compact
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// ---- IO helpers ----

func writeManualFiles(outDir string, files map[string][]byte) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	// deterministic file order
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		p := filepath.Join(outDir, name)
		if err := writeFileDeterministic(p, files[name]); err != nil {
			return err
		}
	}
	return nil
}

func checkManualFiles(outDir string, files map[string][]byte) error {
	// deterministic file order
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		p := filepath.Join(outDir, name)
		existing, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("missing %s (run without --check to generate)", p)
		}
		want := files[name]
		if !bytes.Equal(existing, want) {
			h1 := sha256.Sum256(existing)
			h2 := sha256.Sum256(want)
			return fmt.Errorf("%s out of date (sha256 %x != %x); run `gorp doc gen` to update", p, h1, h2)
		}
	}
	return nil
}

func writeFileDeterministic(path string, b []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp.*.md")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := io.Copy(tmp, bytes.NewReader(b)); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// Windows-friendly replace
	_ = os.Remove(path)
	if err := os.Rename(tmpName, path); err == nil {
		return nil
	}
	// fallback: try direct write
	if err := os.WriteFile(path, b, 0o644); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("write %s: file may be in use (close editors and retry): %w", path, err)
		}
		return err
	}
	return nil
}
