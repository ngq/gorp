package cmd

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	starterTemplateBase         = "base"
	starterTemplateGoLayout     = "golayout"
	starterTemplateGoLayoutWire = "golayout-wire"
	starterTemplateMultiFlat    = "multi-flat"      // 多微服务 - 扁平化
	starterTemplateMultiFlatWire = "multi-flat-wire" // 多微服务 - 扁平化（Wire）
)

var reIdent = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// isValidIdent 判断输入是否满足脚手架命名约束。
func isValidIdent(s string) bool {
	return reIdent.MatchString(s)
}

// toPublicGoName 把常见命名形式转成 Go 导出标识符。
func toPublicGoName(s string) string {
	// snake or kebab -> CamelCase
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}
	return b.String()
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// writeGoFile 会先尝试 gofmt，再把源码写入目标文件。
//
// 中文说明：
// - 如果格式化失败，仍会把原始源码写出，方便开发者排查模板问题。
func writeGoFile(path string, src string) error {
	b, err := format.Source([]byte(src))
	if err != nil {
		// write original for debugging
		b = []byte(src)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func writeTextFile(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func replaceAllInFile(path string, old, new string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := bytes.ReplaceAll(b, []byte(old), []byte(new))
	return os.WriteFile(path, out, 0o644)
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func absJoin(root string, parts ...string) string {
	all := append([]string{root}, parts...)
	return filepath.Join(all...)
}

// requireIdent 校验脚手架输入名称是否合法。
func requireIdent(name, field string) error {
	if !isValidIdent(name) {
		return fmt.Errorf("invalid %s: %s", field, name)
	}
	return nil
}

// validateStarterTemplate 校验 starter 模板名称是否合法。
func validateStarterTemplate(name string) error {
	name = strings.TrimSpace(strings.ToLower(name))
	switch name {
	case "", starterTemplateBase, starterTemplateGoLayout, starterTemplateGoLayoutWire, starterTemplateMultiFlat, starterTemplateMultiFlatWire:
		return nil
	default:
		return fmt.Errorf("unsupported template: %s (supported: base, golayout, golayout-wire, multi-flat, multi-flat-wire)", name)
	}
}

// normalizeStarterTemplate 规范化模板名，空值默认回落到 base。
func normalizeStarterTemplate(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return starterTemplateBase
	}
	return name
}

// validateReleaseStarterTemplate 校验 release 形态当前允许的模板集合。
func validateReleaseStarterTemplate(name string) error {
	name = strings.TrimSpace(strings.ToLower(name))
	switch name {
	case "", starterTemplateBase, starterTemplateGoLayout, starterTemplateGoLayoutWire:
		return nil
	default:
		return fmt.Errorf("unsupported release template: %s (supported: base, golayout, golayout-wire)", name)
	}
}

// resolveOfflineTemplateRoot 返回 offline 模板根目录。
func resolveOfflineTemplateRoot(name string) string {
	switch normalizeStarterTemplate(name) {
	case starterTemplateGoLayout:
		return "templates/golayout/project"
	case starterTemplateGoLayoutWire:
		return "templates/golayout-wire/project"
	case starterTemplateMultiFlat:
		return "templates/multi-flat/project"
	case starterTemplateMultiFlatWire:
		return "templates/multi-flat-wire/project"
	default:
		return "templates/project"
	}
}

// resolveReleaseTemplateRoot 返回 release 模板源目录。
func resolveReleaseTemplateRoot(name string) string {
	switch normalizeStarterTemplate(name) {
	case starterTemplateGoLayout:
		return "templates/release/golayout/project"
	case starterTemplateGoLayoutWire:
		return "templates/release/golayout-wire/project"
	default:
		return "templates/release/project"
	}
}

// defaultReleaseTemplateAsset 返回模板对应的默认 release 资产名。
func defaultReleaseTemplateAsset(name string) string {
	switch normalizeStarterTemplate(name) {
	case starterTemplateGoLayout:
		return "gorp-template-golayout.zip"
	case starterTemplateGoLayoutWire:
		return "gorp-template-golayout-wire.zip"
	default:
		return "gorp-template.zip"
	}
}

// templateFromPreset 根据 preset 名称推断对应的 template。
//
// 中文说明：
// - 当用户只指定 preset 而不显式指定 template 时，自动选择匹配的 template。
// - golayout-basic / golayout-enterprise -> golayout
// - 其他情况返回空字符串，表示使用默认 template。
func templateFromPreset(preset string) string {
	preset = strings.TrimSpace(strings.ToLower(preset))
	switch preset {
	case "golayout-basic", "golayout-enterprise":
		return starterTemplateGoLayout
	default:
		return ""
	}
}

// applyStarterPreset 把用户态 preset 展开为研发态细粒度能力开关。
//
// 中文说明：
// - 用户只需要记住少量"结果导向"的预设；
// - 内部仍保留 `with-*` 细粒度能力，作为模板演进与组合验证的底层能力图谱；
// - preset 只是把"稳定的一组能力组合"映射到这些底层开关上。
func applyStarterPreset(preset string, in *scaffoldInput) error {
	preset = strings.TrimSpace(strings.ToLower(preset))
	if preset == "" {
		return nil
	}
	switch preset {
	case "golayout-basic":
		in.WithDB = true
		in.WithSwagger = true
		in.WithAuth = false
		in.WithRBAC = false
		in.WithAdmin = false
		return nil
	case "golayout-enterprise":
		in.WithDB = true
		in.WithSwagger = true
		in.WithAuth = true
		in.WithRBAC = true
		in.WithAdmin = true
		return nil
	default:
		return fmt.Errorf("unsupported preset: %s (supported: golayout-basic, golayout-enterprise)", preset)
	}
}

// templateDisplayName 返回更适合在 CLI 输出中展示的模板名。
func templateDisplayName(name string) string {
	return normalizeStarterTemplate(name)
}

// validateProjectName 校验脚手架生成目录名，避免路径注入或越界写入。
func validateProjectName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid project name: %s", name)
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("project name must not contain path separators")
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("project name must not be an absolute path")
	}
	return nil
}

// validateModulePath 做最小可用的 Go module 路径校验。
func validateModulePath(mod string) error {
	mod = strings.TrimSpace(mod)
	if mod == "" {
		return fmt.Errorf("module path is required")
	}
	if strings.ContainsAny(mod, " \t\r\n") {
		return fmt.Errorf("module path must not contain whitespace")
	}
	if strings.HasPrefix(mod, "/") || strings.HasSuffix(mod, "/") {
		return fmt.Errorf("invalid module path: %s", mod)
	}
	return nil
}

// validateGitHubRepo 校验 GitHub repo 参数必须是 owner/repo 形式。
func validateGitHubRepo(repo string) error {
	repo = strings.TrimSpace(repo)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("repo must be in <owner>/<repo> format")
	}
	return nil
}

// validateAssetName 校验 release asset 名称。
func validateAssetName(asset string) error {
	asset = strings.TrimSpace(asset)
	if asset == "" {
		return fmt.Errorf("asset is required")
	}
	if strings.Contains(asset, "/") || strings.Contains(asset, "\\") {
		return fmt.Errorf("asset must not contain path separators")
	}
	if !strings.HasSuffix(strings.ToLower(asset), ".zip") {
		return fmt.Errorf("asset must be a .zip file")
	}
	return nil
}

// scaffoldInput 是 offline / from-release 共享的脚手架核心输入。
type scaffoldInput struct {
	Name             string
	Module           string
	FrameworkModule  string
	FrameworkPath    string
	FrameworkVersion string
	Backend          string
	WithDB           bool
	WithSwagger      bool
	WithAuth         bool
	WithRBAC         bool
	WithAdmin        bool
}

func buildScaffoldData(in scaffoldInput) map[string]any {
	return map[string]any{
		"Name":             in.Name,
		"Module":           in.Module,
		"ModuleName":       in.Module, // 别名，供模板使用
		"FrameworkModule":  in.FrameworkModule,
		"FrameworkPath":    normalizeFrameworkReplacePath(in.FrameworkPath),
		"FrameworkVersion": in.FrameworkVersion,
		"Backend":          in.Backend,
		"IsGormBackend":    in.Backend == "" || in.Backend == "gorm",
		"IsEntBackend":     in.Backend == "ent",
		"WithDB":           in.WithDB,
		"WithSwagger":      in.WithSwagger,
		"WithAuth":         in.WithAuth,
		"WithRBAC":         in.WithRBAC,
		"WithAdmin":        in.WithAdmin,
	}
}

func normalizeFrameworkReplacePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	// 中文说明：
	// - 在 Windows + bash/msys 环境下，用户可能输入 `/e/project/...` 这种类 Unix 盘符路径；
	// - Go 的 filepath.Abs 在 Windows 下会把它当作相对路径处理，产生 `E:/e/...` 这类错误结果；
	// - 这里先把 `/x/...` 规范化成 `X:/...`，再做 Clean/Abs，保证生成到 go.mod replace 的路径是本机可识别绝对路径。
	if len(path) >= 3 && path[0] == '/' && path[2] == '/' {
		drive := path[1]
		if (drive >= 'a' && drive <= 'z') || (drive >= 'A' && drive <= 'Z') {
			path = strings.ToUpper(string(drive)) + ":" + path[2:]
		}
	}

	cleaned := filepath.Clean(path)
	if abs, err := filepath.Abs(cleaned); err == nil {
		cleaned = abs
	}
	return filepath.ToSlash(cleaned)
}

func promptProjectInput(r *bufio.Reader, out io.Writer, defaultFrameworkPath string, needFrameworkPath bool) (scaffoldInput, string, error) {
	name, err := promptStringR(r, out, "请输入目录名称：", "", true)
	if err != nil {
		return scaffoldInput{}, "", err
	}
	name = strings.TrimSpace(name)
	if err := validateProjectName(name); err != nil {
		return scaffoldInput{}, "", err
	}

	mod, err := promptStringR(r, out, "请输入模块名称(go.mod中的module, 默认为文件夹名称)：", name, false)
	if err != nil {
		return scaffoldInput{}, "", err
	}
	mod = strings.TrimSpace(mod)
	if mod == "" {
		mod = name
	}
	if err := validateModulePath(mod); err != nil {
		return scaffoldInput{}, "", err
	}

	in := scaffoldInput{
		Name:             name,
		Module:           mod,
		FrameworkModule:  "github.com/ngq/gorp",
		FrameworkVersion: "v0.0.0",
	}
	if needFrameworkPath {
		frameworkPath, err := promptStringR(r, out, "请输入框架源码路径(用于 go.mod replace，默认当前目录)：", defaultFrameworkPath, false)
		if err != nil {
			return scaffoldInput{}, "", err
		}
		frameworkPath = strings.TrimSpace(frameworkPath)
		if frameworkPath == "" {
			frameworkPath = defaultFrameworkPath
		}
		in.FrameworkPath = frameworkPath
	}
	return in, name, nil
}

func prepareScaffoldTargetDir(baseDir, projectName string) (string, error) {
	target := filepath.Join(baseDir, projectName)
	if !dirExists(target) {
		return target, nil
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return "", err
	}
	if len(entries) > 0 {
		return "", fmt.Errorf("target folder already exists and is not empty: %s", target)
	}
	return target, nil
}

func renderTemplateProject(src fs.FS, srcRoot, dstRoot string, data map[string]any) error {
	return renderTemplateDir(src, srcRoot, dstRoot, data)
}

func printScaffoldNext(out io.Writer, folder string) {
	fmt.Fprintf(out, "created: %s\n", folder)
	fmt.Fprintln(out, "next: cd", folder)
	fmt.Fprintln(out, "      go mod tidy")
	fmt.Fprintln(out, "      go run ./cmd/app")
}

func buildGitHubReleaseAssetURL(repo, tag, asset string) string {
	if strings.EqualFold(tag, "latest") {
		return fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", repo, asset)
	}
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, asset)
}

func downloadReleaseAsset(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download asset failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

type zipFS struct {
	files map[string][]byte
	dirs  map[string]struct{}
}

func newZipFSFromBytes(b []byte) (fs.FS, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}
	zfs := &zipFS{files: map[string][]byte{}, dirs: map[string]struct{}{".": {}}}
	for _, f := range zr.File {
		name := filepath.ToSlash(strings.TrimSpace(f.Name))
		if name == "" {
			continue
		}
		if strings.HasPrefix(name, "/") || strings.Contains(name, "../") || strings.HasPrefix(name, "../") {
			return nil, fmt.Errorf("invalid zip entry path: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			zfs.addDir(strings.TrimSuffix(name, "/"))
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		zfs.files[name] = content
		zfs.addParents(name)
	}
	return zfs, nil
}

func (z *zipFS) addDir(name string) {
	name = strings.Trim(name, "/")
	if name == "" {
		z.dirs["."] = struct{}{}
		return
	}
	z.dirs[name] = struct{}{}
	z.addParents(name)
}

func (z *zipFS) addParents(name string) {
	dir := filepath.ToSlash(filepath.Dir(name))
	for dir != "." && dir != "/" && dir != "" {
		z.dirs[dir] = struct{}{}
		next := filepath.ToSlash(filepath.Dir(dir))
		if next == dir {
			break
		}
		dir = next
	}
	z.dirs["."] = struct{}{}
}

func (z *zipFS) Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(filepath.ToSlash(name), "./")
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		name = "."
	}
	if b, ok := z.files[name]; ok {
		return &zipMemFile{reader: bytes.NewReader(b), info: zipFileInfo{name: filepath.Base(name), size: int64(len(b))}}, nil
	}
	if _, ok := z.dirs[name]; ok {
		return z.openDir(name)
	}
	return nil, fs.ErrNotExist
}

func (z *zipFS) openDir(name string) (fs.File, error) {
	items := make([]fs.DirEntry, 0)
	prefix := ""
	if name != "." {
		prefix = name + "/"
	}
	seen := map[string]struct{}{}
	for dir := range z.dirs {
		if dir == name || !strings.HasPrefix(dir, prefix) {
			continue
		}
		rest := strings.TrimPrefix(dir, prefix)
		if rest == "" || strings.Contains(rest, "/") {
			continue
		}
		if _, ok := seen[rest]; ok {
			continue
		}
		seen[rest] = struct{}{}
		items = append(items, zipDirEntry{info: zipFileInfo{name: rest, dir: true}})
	}
	for file := range z.files {
		if !strings.HasPrefix(file, prefix) {
			continue
		}
		rest := strings.TrimPrefix(file, prefix)
		if rest == "" || strings.Contains(rest, "/") {
			continue
		}
		if _, ok := seen[rest]; ok {
			continue
		}
		seen[rest] = struct{}{}
		items = append(items, zipDirEntry{info: zipFileInfo{name: rest, size: int64(len(z.files[file]))}})
	}
	return &zipDirFile{info: zipFileInfo{name: filepath.Base(name), dir: true}, entries: items}, nil
}

type zipMemFile struct {
	reader *bytes.Reader
	info   zipFileInfo
}

func (f *zipMemFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *zipMemFile) Read(p []byte) (int, error) { return f.reader.Read(p) }
func (f *zipMemFile) Close() error               { return nil }

type zipDirFile struct {
	info    zipFileInfo
	entries []fs.DirEntry
	offset  int
}

func (f *zipDirFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *zipDirFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (f *zipDirFile) Close() error               { return nil }
func (f *zipDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.offset >= len(f.entries) && n > 0 {
		return nil, io.EOF
	}
	if n <= 0 || f.offset+n > len(f.entries) {
		n = len(f.entries) - f.offset
	}
	out := f.entries[f.offset : f.offset+n]
	f.offset += n
	return out, nil
}

type zipDirEntry struct{ info zipFileInfo }

func (d zipDirEntry) Name() string               { return d.info.Name() }
func (d zipDirEntry) IsDir() bool                { return d.info.IsDir() }
func (d zipDirEntry) Type() fs.FileMode          { return d.info.Mode().Type() }
func (d zipDirEntry) Info() (fs.FileInfo, error) { return d.info, nil }

type zipFileInfo struct {
	name string
	size int64
	dir  bool
}

func (i zipFileInfo) Name() string       { return i.name }
func (i zipFileInfo) Size() int64        { return i.size }
func (i zipFileInfo) Mode() fs.FileMode  { if i.dir { return fs.ModeDir | 0o755 }; return 0o644 }
func (i zipFileInfo) ModTime() time.Time { return time.Time{} }
func (i zipFileInfo) IsDir() bool        { return i.dir }
func (i zipFileInfo) Sys() any           { return nil }

func zipDirectoryFromFS(src fs.FS, srcRoot string, dstZip string, zipRoot string) error {
	if err := os.MkdirAll(filepath.Dir(dstZip), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dstZip)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	paths := make([]string, 0)
	if err := fs.WalkDir(src, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(paths)

	zipRoot = strings.Trim(filepath.ToSlash(zipRoot), "/")
	for _, path := range paths {
		rel := strings.TrimPrefix(path, srcRoot)
		rel = strings.TrimPrefix(rel, "/")
		rel = strings.TrimPrefix(rel, "\\")
		entryName := zipRoot
		if rel != "" {
			entryName = filepath.ToSlash(filepath.Join(zipRoot, rel))
		}
		info, err := fs.Stat(src, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if rel == "" {
				continue
			}
			_, err := zw.Create(strings.TrimSuffix(entryName, "/") + "/")
			if err != nil {
				return err
			}
			continue
		}
		w, err := zw.Create(entryName)
		if err != nil {
			return err
		}
		b, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		if _, err := w.Write(b); err != nil {
			return err
		}
	}
	return nil
}

// promptString 从 stdin 读取一行文本（非 survey，兼容 Windows/MSYS 终端环境）。
//
// 重要说明（易踩坑）：
// - 如果一个命令需要连续询问多个问题（多次读取 stdin），不要在每次询问时都 new bufio.Reader。
//   因为 bufio.Reader 可能会"预读"后续输入并缓存；如果换了新的 Reader，会导致后续输入丢失。
// - 解决方案：在命令开始时创建一个 reader，然后调用 promptStringR/promptConfirmR。
//
// - 如果用户直接回车且 defaultValue 非空，则返回 defaultValue
// - 如果 required=true，会一直提示直到输入非空
func promptString(in io.Reader, out io.Writer, message string, defaultValue string, required bool) (string, error) {
	r := bufio.NewReader(in)
	return promptStringR(r, out, message, defaultValue, required)
}

// promptStringR 是 promptString 的 reader 复用版本。
func promptStringR(r *bufio.Reader, out io.Writer, message string, defaultValue string, required bool) (string, error) {
	for {
		if defaultValue != "" {
			fmt.Fprintf(out, "%s(默认: %s): ", message, defaultValue)
		} else {
			fmt.Fprintf(out, "%s", message)
			if !strings.HasSuffix(message, ":") && !strings.HasSuffix(message, "：") {
				fmt.Fprint(out, ": ")
			} else {
				fmt.Fprint(out, " ")
			}
		}
		line, err := r.ReadString('\n')
		if err != nil {
			// EOF 也视为一次输入
			if err == io.EOF {
				line = strings.TrimRight(line, "\r\n")
			} else {
				return "", err
			}
		}
		line = strings.TrimSpace(line)
		if line == "" && defaultValue != "" {
			line = defaultValue
		}
		if required && strings.TrimSpace(line) == "" {
			fmt.Fprintln(out, "输入不能为空，请重新输入。")
			continue
		}
		return line, nil
	}
}

// promptConfirm 读取 y/N 确认。
func promptConfirm(in io.Reader, out io.Writer, message string, defaultYes bool) (bool, error) {
	r := bufio.NewReader(in)
	return promptConfirmR(r, out, message, defaultYes)
}

// promptConfirmR 是 promptConfirm 的 reader 复用版本。
func promptConfirmR(r *bufio.Reader, out io.Writer, message string, defaultYes bool) (bool, error) {
	def := "N"
	if defaultYes {
		def = "Y"
	}
	ans, err := promptStringR(r, out, fmt.Sprintf("%s (y/N)", message), def, true)
	if err != nil {
		return false, err
	}
	ans = strings.ToLower(strings.TrimSpace(ans))
	if ans == "y" || ans == "yes" {
		return true, nil
	}
	if ans == "n" || ans == "no" {
		return false, nil
	}
	// 默认值兜底
	return defaultYes, nil
}
