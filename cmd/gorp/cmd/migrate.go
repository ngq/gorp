package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"    // mysql 驱动（golang-migrate 按需加载）
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres 驱动
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"  // sqlite3 驱动
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source（从本地文件系统加载迁移文件）
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// dbConfig 对应 config/app.yaml 中的 database 配置段。
//
// 中文说明：
// - driver：数据库驱动类型，支持 sqlite / mysql / postgres；
// - dsn：数据库连接字符串，格式因驱动不同而异：
//   - sqlite：  "file:demo.db?cache=shared&mode=rwc"
//   - mysql：   "user:password@tcp(127.0.0.1:3306)/dbname?parseTime=true"
//   - postgres："postgres://user:password@localhost:5432/dbname?sslmode=disable"
type dbConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// appYamlConfig 对应 config/app.yaml 的顶层结构。
//
// 中文说明：
// - 只解析 database 段，其余配置段不做展开；
// - 如果 database 段不存在或字段为空，migrate 命令会给出明确的错误提示。
type appYamlConfig struct {
	Database *dbConfig `yaml:"database"`
}

// migrateCmd 是数据库版本化迁移的命令组。
//
// 中文说明：
// - 提供数据库迁移的创建、执行、回滚能力；
// - 子命令：
//   1. `gorp migrate create <name>`：创建一对 up/down 迁移文件；
//   2. `gorp migrate up`：执行所有待执行的迁移；
//   3. `gorp migrate down`：回滚最后一个迁移；
// - 迁移文件存放在项目根目录的 migrations/ 目录下。
var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "数据库版本化迁移",
	GroupID: commandGroupAdvanced,
	Long: `数据库版本化迁移（基于 golang-migrate）。

子命令：
  gorp migrate create <name>  创建迁移文件（生成 up/down 一对 SQL 文件）
  gorp migrate up             执行所有待执行的迁移
  gorp migrate down           回滚最后一个迁移

迁移文件存放在项目根目录的 migrations/ 目录下。
数据库配置从 config/app.yaml 的 database 段读取。`,
}

// migrateCreateCmd 创建一对迁移文件。
//
// 中文说明：
// - 在 migrations/ 目录下生成 <序号>_<名称>.up.sql 和 <序号>_<名称>.down.sql；
// - 序号自动递增，基于目录中已有文件的最大序号 +1；
// - 序号格式为 6 位数字（000001, 000002, ...），与 golang-migrate 约定一致。
var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "创建迁移文件",
	Long: `创建迁移文件。

在 migrations/ 目录下生成一对 up/down SQL 文件。
序号自动递增，例如：
  migrations/000001_add_users_table.up.sql
  migrations/000001_add_users_table.down.sql`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrateCreate,
}

// migrateUpCmd 执行所有待执行的迁移。
//
// 中文说明：
// - 从 config/app.yaml 读取 database.driver 和 database.dsn；
// - 打开数据库连接并应用 migrations/ 目录下所有未执行的迁移；
// - 如果没有待执行的迁移，会提示"no new migrations"。
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行所有待执行的迁移",
	Long: `执行所有待执行的迁移。

从 config/app.yaml 的 database 段读取数据库配置，
然后应用 migrations/ 目录下所有未执行的迁移。`,
	RunE: runMigrateUp,
}

// migrateDownCmd 回滚最后一个迁移。
//
// 中文说明：
// - 从 config/app.yaml 读取数据库配置；
// - 回滚最近一次已执行的迁移（即 -1 步）。
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚最后一个迁移",
	Long: `回滚最后一个迁移。

从 config/app.yaml 的 database 段读取数据库配置，
然后回滚最近一次已执行的迁移。`,
	RunE: runMigrateDown,
}

// reMigrationFile 匹配迁移文件名中的序号部分。
//
// 中文说明：
// - golang-migrate 约定文件名格式为 <序号>_<名称>.up.sql 或 <序号>_<名称>.down.sql；
// - 序号可以是纯数字（如 000001）或时间戳格式（如 20260101120000）；
// - 此正则提取序号部分，用于自动递增。
var reMigrationFile = regexp.MustCompile(`^(\d+)_.*\.sql$`)

// migrationsDir 返回迁移文件的目录路径。
//
// 中文说明：
// - 固定为当前工作目录下的 migrations/；
// - 如果目录不存在，由调用方决定是否创建。
func migrationsDir() string {
	return filepath.Join(".", "migrations")
}

// nextMigrationVersion 扫描 migrations/ 目录，返回下一个可用的序号。
//
// 中文说明：
// - 读取目录中所有已有迁移文件，提取最大序号；
// - 最大序号 +1 即为下一个版本号；
// - 如果目录为空或不存在，返回 1。
func nextMigrationVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxVersion := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := reMigrationFile.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}
		v, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if v > maxVersion {
			maxVersion = v
		}
	}
	return maxVersion + 1, nil
}

// formatMigrationVersion 将版本号格式化为 6 位字符串。
//
// 中文说明：
// - golang-migrate 约定序号部分为定长数字字符串；
// - 6 位可以支持到 999999 个迁移，足够大多数项目使用。
func formatMigrationVersion(v int) string {
	return fmt.Sprintf("%06d", v)
}

// sanitizeMigrationName 将用户输入的迁移名称转为合法的文件名片段。
//
// 中文说明：
// - 将空格和连续连字符/下划线替换为单个下划线；
// - 转为小写，避免文件名大小写不一致问题；
// - 去除首尾的下划线和连字符。
func sanitizeMigrationName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// 空格和连字符替换为下划线
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	// 去除连续下划线
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	// 去除首尾下划线
	name = strings.Trim(name, "_")
	if name == "" {
		name = "migration"
	}
	return name
}

// runMigrateCreate 创建一对迁移文件。
//
// 中文说明：
// - 1) 确保 migrations/ 目录存在；
// - 2) 计算下一个可用序号；
// - 3) 生成 .up.sql 和 .down.sql 两个空文件；
// - 4) 输出文件路径，方便用户编辑。
func runMigrateCreate(cmd *cobra.Command, args []string) error {
	name := sanitizeMigrationName(args[0])
	dir := migrationsDir()

	// 确保迁移目录存在
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建迁移目录失败：%w", err)
	}

	// 计算下一个版本号
	version, err := nextMigrationVersion(dir)
	if err != nil {
		return fmt.Errorf("扫描迁移文件失败：%w", err)
	}

	versionStr := formatMigrationVersion(version)
	upFile := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", versionStr, name))
	downFile := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", versionStr, name))

	// 写入空文件（golang-migrate 要求文件存在且非空，写入占位注释）
	upContent := fmt.Sprintf("-- +migrate Up\n-- 在此编写向上迁移的 SQL\n")
	downContent := fmt.Sprintf("-- +migrate Down\n-- 在此编写回滚迁移的 SQL\n")

	if err := os.WriteFile(upFile, []byte(upContent), 0o644); err != nil {
		return fmt.Errorf("创建迁移文件失败：%w", err)
	}
	if err := os.WriteFile(downFile, []byte(downContent), 0o644); err != nil {
		return fmt.Errorf("创建迁移文件失败：%w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "已创建迁移文件：\n  %s\n  %s\n", upFile, downFile)
	return nil
}

// loadDBConfig 从 config/app.yaml 读取数据库配置。
//
// 中文说明：
// - 读取项目根目录下的 config/app.yaml；
// - 解析 database 段的 driver 和 dsn 字段；
// - 如果配置文件不存在或 database 段缺失/字段为空，返回明确错误。
func loadDBConfig() (*dbConfig, error) {
	configPath := filepath.Join(".", "config", "app.yaml")
	if !fileExists(configPath) {
		return nil, fmt.Errorf("未找到配置文件：%s\n请确保在 gorp 项目根目录下运行此命令", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败：%w", err)
	}

	var cfg appYamlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%w", err)
	}

	if cfg.Database == nil {
		return nil, fmt.Errorf("配置文件 %s 中缺少 database 配置段\n请参考以下格式添加：\n\ndatabase:\n  driver: \"sqlite\"\n  dsn: \"file:demo.db?cache=shared&mode=rwc\"", configPath)
	}

	if cfg.Database.Driver == "" {
		return nil, fmt.Errorf("配置文件中 database.driver 为空\n请指定 driver（支持：sqlite / mysql / postgres）")
	}

	if cfg.Database.DSN == "" {
		return nil, fmt.Errorf("配置文件中 database.dsn 为空\n请指定数据库连接字符串")
	}

	return cfg.Database, nil
}

// buildMigrateDSN 根据 driver 类型构建 golang-migrate 所需的 DSN 格式。
//
// 中文说明：
// - golang-migrate 对不同数据库的 DSN 格式有不同要求；
// - sqlite3：直接使用 dsn 字符串；
// - mysql：直接使用 dsn 字符串；
// - postgres：直接使用 dsn 字符串；
// - 返回格式为 "<driver>://<dsn>"，供 migrate.New() 使用。
func buildMigrateDSN(cfg *dbConfig) (string, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	dsn := strings.TrimSpace(cfg.DSN)

	switch driver {
	case "sqlite", "sqlite3":
		// golang-migrate 的 sqlite3 驱动需要 "sqlite3://" 前缀
		return "sqlite3://" + dsn, nil
	case "mysql":
		// golang-migrate 的 mysql 驱动需要 "mysql://" 前缀
		return "mysql://" + dsn, nil
	case "postgres", "postgresql":
		// golang-migrate 的 postgres 驱动支持 postgres:// 格式
		// 如果 dsn 已经以 postgres:// 开头则不再添加前缀
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			return dsn, nil
		}
		return dsn, nil
	default:
		return "", fmt.Errorf("不支持的数据库驱动：%s（支持：sqlite / mysql / postgres）", driver)
	}
}

// newMigrateInstance 创建 golang-migrate 实例。
//
// 中文说明：
// - 从配置文件读取数据库连接信息；
// - 从 migrations/ 目录加载迁移文件源；
// - 返回可用于 up/down 操作的 migrate.Migrate 实例。
func newMigrateInstance() (*migrate.Migrate, error) {
	cfg, err := loadDBConfig()
	if err != nil {
		return nil, err
	}

	dir := migrationsDir()
	if !dirExists(dir) {
		return nil, fmt.Errorf("迁移目录不存在：%s\n请先使用 gorp migrate create <name> 创建迁移文件", dir)
	}

	// 检查迁移目录中是否有文件
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败：%w", err)
	}
	sqlFiles := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles++
		}
	}
	if sqlFiles == 0 {
		return nil, fmt.Errorf("迁移目录 %s 中没有迁移文件\n请先使用 gorp migrate create <name> 创建迁移文件", dir)
	}

	dsn, err := buildMigrateDSN(cfg)
	if err != nil {
		return nil, err
	}

	// 获取迁移文件的绝对路径（golang-migrate 的 file source 需要绝对路径或 file:// 前缀）
	migrationsAbs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("获取迁移目录绝对路径失败：%w", err)
	}

	m, err := migrate.New(
		"file://"+filepath.ToSlash(migrationsAbs),
		dsn,
	)
	if err != nil {
		return nil, fmt.Errorf("初始化迁移引擎失败：%w\n请检查数据库配置是否正确（driver=%s, dsn=%s）", err, cfg.Driver, cfg.DSN)
	}

	return m, nil
}

// runMigrateUp 执行所有待执行的迁移。
//
// 中文说明：
// - 创建 migrate 实例并调用 Up()；
// - 如果没有待执行的迁移，输出提示信息；
// - 其他错误（如连接失败、SQL 语法错误）会原样返回。
func runMigrateUp(cmd *cobra.Command, args []string) error {
	m, err := newMigrateInstance()
	if err != nil {
		return err
	}
	defer m.Close()

	fmt.Fprintln(cmd.OutOrStdout(), "正在执行迁移...")
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Fprintln(cmd.OutOrStdout(), "没有待执行的迁移（已是最新版本）")
			return nil
		}
		return fmt.Errorf("执行迁移失败：%w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "迁移执行完成")
	return nil
}

// runMigrateDown 回滚最后一个迁移。
//
// 中文说明：
// - 创建 migrate 实例并调用 Steps(-1)；
// - 回滚最近一次已执行的迁移；
// - 如果没有可回滚的迁移，输出提示信息。
func runMigrateDown(cmd *cobra.Command, args []string) error {
	m, err := newMigrateInstance()
	if err != nil {
		return err
	}
	defer m.Close()

	fmt.Fprintln(cmd.OutOrStdout(), "正在回滚最后一个迁移...")
	if err := m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Fprintln(cmd.OutOrStdout(), "没有可回滚的迁移")
			return nil
		}
		return fmt.Errorf("回滚迁移失败：%w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "回滚完成")
	return nil
}

// listMigrationFiles 列出 migrations/ 目录中的迁移文件（供未来扩展使用）。
//
// 中文说明：
// - 按文件名排序，返回所有 .sql 文件；
// - 仅供内部使用，当前未对外暴露为子命令。
func listMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
}
