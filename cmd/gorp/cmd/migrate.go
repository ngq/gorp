package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ngq/gorp/framework/pkg/migrate"
)

// migrateFlags 存储 migrate 命令全局标志
var migrateFlags = struct {
	dir       string
	config    string
	tableName string
}{
	dir:       "migrations",
	config:    "config/app.yaml",
	tableName: goose.DefaultTablename,
}

// dbConfig 对应 config/app.yaml 中的 database 配置段。
type dbConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// appYamlConfig 对应 config/app.yaml 的顶层结构。
type appYamlConfig struct {
	Database *dbConfig `yaml:"database"`
}

// migrateCmd 是数据库版本化迁移的根命令。
var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "数据库版本化迁移（基于 goose）",
	GroupID: commandGroupAdvanced,
	Long: `数据库版本化迁移（基于 pressly/goose/v3）。

提供完整的数据库迁移管理能力，支持单文件 Up/Down 事务包装与 Go 代码迁移。

子命令：
  gorp migrate status                 查看迁移应用状态
  gorp migrate create <name> [sql|go] 创建迁移文件（默认生成方言自适应的 SQL 模板）
  gorp migrate up                     执行所有待执行的迁移
  gorp migrate up-by-one              仅执行下一个未应用的迁移
  gorp migrate up-to <version>        迁移到指定版本
  gorp migrate down                   回滚最后一个迁移
  gorp migrate down-to <version>      回滚到指定版本
  gorp migrate redo                   回滚上一个迁移并重新执行（开发调试神器）
  gorp migrate reset                  回滚所有已应用的迁移
  gorp migrate version                查看当前数据库版本`,
}

// migrateStatusCmd 查看迁移状态
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	Long:  "查看所有迁移文件的应用状态及执行时间。",
	RunE:  runMigrateStatus,
}

// migrateCreateCmd 创建迁移文件
var migrateCreateCmd = &cobra.Command{
	Use:   "create <name> [sql|go]",
	Short: "创建迁移文件",
	Long: `创建迁移文件。

默认在 migrations/ 目录下生成基于时间戳命名的单文件迁移脚本：
  migrations/YYYYMMDDHHMMSS_<name>.sql

同时会根据 config/app.yaml 中配置的数据库驱动（MySQL / PostgreSQL / SQLite）动态匹配方言骨架模板。`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runMigrateCreate,
}

// migrateUpCmd 执行全量迁移
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行所有待执行的迁移",
	Long:  "从配置文件读取数据库配置，并应用 migrations/ 目录下所有未执行的迁移。",
	RunE:  runMigrateUp,
}

// migrateUpByOneCmd 单步执行迁移
var migrateUpByOneCmd = &cobra.Command{
	Use:   "up-by-one",
	Short: "仅执行下一个待执行的迁移",
	Long:  "仅向前应用一个版本的迁移，适合精细化验证。",
	RunE:  runMigrateUpByOne,
}

// migrateUpToCmd 迁移到指定版本
var migrateUpToCmd = &cobra.Command{
	Use:   "up-to <version>",
	Short: "迁移到指定版本",
	Long:  "应用迁移直至达到指定的版本号（支持向前或追平指定版本）。",
	Args:  cobra.ExactArgs(1),
	RunE:  runMigrateUpTo,
}

// migrateDownCmd 回滚最后一个迁移
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚最后一个迁移",
	Long:  "回滚最近一次已执行的迁移版本。",
	RunE:  runMigrateDown,
}

// migrateDownToCmd 回滚到指定版本
var migrateDownToCmd = &cobra.Command{
	Use:   "down-to <version>",
	Short: "回滚到指定版本",
	Long:  "连续回滚迁移直至当前数据库版本降至指定版本。",
	Args:  cobra.ExactArgs(1),
	RunE:  runMigrateDownTo,
}

// migrateRedoCmd 重跑上一个迁移
var migrateRedoCmd = &cobra.Command{
	Use:   "redo",
	Short: "回滚上一个迁移并重新执行",
	Long:  "回滚最近应用的迁移并立即重新执行（Down + Up），非常适合本地反复调试 SQL。",
	RunE:  runMigrateRedo,
}

// migrateResetCmd 回滚所有迁移
var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "回滚所有已执行的迁移",
	Long:  "回滚所有已应用的迁移至初始状态（警告：这将清空所有表结构，仅限测试/开发使用）。",
	RunE:  runMigrateReset,
}

// migrateVersionCmd 查看当前数据库版本
var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "查看当前数据库版本",
	Long:  "查询当前数据库版本记录表，输出最新已应用的迁移版本号。",
	RunE:  runMigrateVersion,
}

// sanitizeMigrationName 将用户输入的迁移名称转为合法的文件名片段。
func sanitizeMigrationName(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' || r == '/' || r == '\\' {
			b.WriteRune('_')
		}
	}
	res := b.String()
	for strings.Contains(res, "__") {
		res = strings.ReplaceAll(res, "__", "_")
	}
	res = strings.Trim(res, "_")
	if res == "" {
		return "migration"
	}
	return res
}

// commandContext 返回 cmd 的上下文，若未设置或为 nil 则回退到 context.Background()。
func commandContext(cmd *cobra.Command) context.Context {
	if cmd != nil && cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}

// loadDBConfig 从指定的配置文件路径读取数据库配置。
func loadDBConfig(configPath string) (*dbConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件不存在：%s\n请确认当前工作目录是否正确，或通过 --config 指定配置文件", configPath)
		}
		return nil, fmt.Errorf("读取配置文件失败：%w", err)
	}

	var appCfg appYamlConfig
	if err := yaml.Unmarshal(data, &appCfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%w", err)
	}

	if appCfg.Database == nil {
		return nil, fmt.Errorf("配置文件 %s 中缺少 database 配置段", configPath)
	}

	driver := strings.TrimSpace(appCfg.Database.Driver)
	dsn := strings.TrimSpace(appCfg.Database.DSN)

	if driver == "" {
		return nil, fmt.Errorf("配置文件 %s 中的 database.driver 不能为空（支持：mysql / postgres / sqlite）", configPath)
	}
	if dsn == "" {
		return nil, fmt.Errorf("配置文件 %s 中的 database.dsn 不能为空", configPath)
	}

	return &dbConfig{
		Driver: driver,
		DSN:    dsn,
	}, nil
}

// normalizeSQLDriver 将驱动别名转换为标准 database/sql 驱动名称。
func normalizeSQLDriver(driver string) (string, error) {
	d := strings.ToLower(strings.TrimSpace(driver))
	switch d {
	case "mysql":
		return "mysql", nil
	case "postgres", "postgresql", "pgsql", "pgx":
		return "pgx", nil
	case "sqlite", "sqlite3":
		return "sqlite3", nil
	default:
		return "", fmt.Errorf("不支持的数据库驱动：%s（支持：mysql / postgres / sqlite）", driver)
	}
}

// openDB 创建数据库连接并进行连通性检查。
func openDB(cfg *dbConfig) (*sql.DB, error) {
	stdDriver, err := normalizeSQLDriver(cfg.Driver)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(stdDriver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败：%w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接检查失败（driver=%s）：%w", cfg.Driver, err)
	}

	return db, nil
}

// getProvider 构建隔离的 goose.Provider 实例。
func getProvider(cfg *dbConfig, db *sql.DB) (*goose.Provider, error) {
	dir := filepath.Clean(migrateFlags.dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("迁移目录不存在：%s\n请先使用 gorp migrate create <name> 创建迁移文件", dir)
	}

	return migrate.NewProvider(
		db,
		cfg.Driver,
		os.DirFS(dir),
		".",
		migrate.WithTableName(migrateFlags.tableName),
	)
}

// generateSQLTemplate 根据当前数据库驱动生成自适应方言的 SQL 初始骨架。
func generateSQLTemplate(driver, tableName string) string {
	d := strings.ToLower(strings.TrimSpace(driver))
	cleanTable := sanitizeMigrationName(tableName)

	switch d {
	case "postgres", "postgresql", "pgsql", "pgx":
		return fmt.Sprintf(`-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "%s" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "%s";
-- +goose StatementEnd
`, cleanTable, cleanTable)

	case "sqlite", "sqlite3":
		return fmt.Sprintf(`-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "%s" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "%s";
-- +goose StatementEnd
`, cleanTable, cleanTable)

	default: // MySQL default
		return fmt.Sprintf(`-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS `+"`%s`"+` (
    `+"`id`"+` BIGINT NOT NULL AUTO_INCREMENT,
    `+"`name`"+` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '名称',
    `+"`created_at`"+` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `+"`updated_at`"+` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`+"`id`"+`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='%s 表';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `+"`%s`"+`;
-- +goose StatementEnd
`, cleanTable, cleanTable, cleanTable)
	}
}

// generateGoTemplate 生成 Go 代码迁移模板
func generateGoTemplate(version int64, name string) string {
	return fmt.Sprintf(`package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up%d, down%d)
}

func up%d(ctx context.Context, tx *sql.Tx) error {
	// 在此编写向上迁移的业务/数据库操作逻辑
	fmt.Println("Applying Go migration %s (%d)...")
	return nil
}

func down%d(ctx context.Context, tx *sql.Tx) error {
	// 在此编写回滚操作逻辑
	fmt.Println("Rolling back Go migration %s (%d)...")
	return nil
}
`, version, version, version, name, version, version, name, version)
}

// runMigrateCreate 创建迁移文件。
func runMigrateCreate(cmd *cobra.Command, args []string) error {
	name := sanitizeMigrationName(args[0])
	format := "sql"
	if len(args) > 1 {
		format = strings.ToLower(strings.TrimSpace(args[1]))
	}
	if format != "sql" && format != "go" {
		return fmt.Errorf("不支持的迁移格式：%s（仅支持 sql 或 go）", format)
	}

	dir := filepath.Clean(migrateFlags.dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建迁移目录失败：%w", err)
	}

	// 生成 14 位 UTC 时间戳版本号，避免多时区协作冲突
	timestampStr := time.Now().UTC().Format("20060102150405")
	versionNum, _ := strconv.ParseInt(timestampStr, 10, 64)

	fileName := fmt.Sprintf("%s_%s.%s", timestampStr, name, format)
	filePath := filepath.Join(dir, fileName)

	// 探测 driver 用于生成匹配的模板（若配置文件读取失败则使用通用 MySQL/SQL 模板）
	driver := "mysql"
	if cfg, err := loadDBConfig(migrateFlags.config); err == nil && cfg != nil {
		driver = cfg.Driver
	}

	var content string
	if format == "go" {
		content = generateGoTemplate(versionNum, name)
	} else {
		content = generateSQLTemplate(driver, name)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入迁移文件失败：%w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "成功创建迁移文件: %s\n", filePath)
	return nil
}

// runMigrateStatus 查看迁移状态
func runMigrateStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		if errors.Is(err, goose.ErrNoMigrations) {
			fmt.Fprintln(cmd.OutOrStdout(), "迁移目录下未找到任何迁移文件，请使用 gorp migrate create 创建")
			return nil
		}
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 15*time.Second)
	defer cancel()

	statuses, err := provider.Status(ctx)
	if err != nil {
		return fmt.Errorf("获取迁移状态失败：%w", err)
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "VERSION\tSTATUS\tAPPLIED AT\tSOURCE")
	fmt.Fprintln(w, "-------\t------\t----------\t------")

	for _, s := range statuses {
		appliedAt := "-"
		if !s.AppliedAt.IsZero() {
			appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
		}
		statusStr := "Pending"
		if s.State == goose.StateApplied {
			statusStr = "Applied"
		}
		version := int64(0)
		sourceName := "-"
		if s.Source != nil {
			version = s.Source.Version
			sourceName = filepath.Base(s.Source.Path)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", version, statusStr, appliedAt, sourceName)
	}
	w.Flush()
	return nil
}

// runMigrateUp 执行全量迁移
func runMigrateUp(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		if errors.Is(err, goose.ErrNoMigrations) {
			fmt.Fprintln(cmd.OutOrStdout(), "迁移目录下未找到任何迁移文件，请使用 gorp migrate create 创建")
			return nil
		}
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 60*time.Second)
	defer cancel()

	fmt.Fprintln(cmd.OutOrStdout(), "正在执行迁移...")
	results, err := provider.Up(ctx)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "没有待执行的迁移（已是最新版本）")
			return nil
		}
		return fmt.Errorf("执行迁移失败：%w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "没有待执行的迁移（已是最新版本）")
		return nil
	}

	for _, res := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "-> 已应用版本 %d (%s) 耗时: %v\n", res.Source.Version, filepath.Base(res.Source.Path), res.Duration)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "迁移执行完成")
	return nil
}

// runMigrateUpByOne 执行下一个迁移
func runMigrateUpByOne(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 30*time.Second)
	defer cancel()

	res, err := provider.UpByOne(ctx)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "没有待执行的迁移（已是最新版本）")
			return nil
		}
		if errors.Is(err, goose.ErrNoMigrations) {
			fmt.Fprintln(cmd.OutOrStdout(), "迁移目录下未找到任何迁移文件，请使用 gorp migrate create 创建")
			return nil
		}
		return fmt.Errorf("执行单步迁移失败：%w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "-> 已应用版本 %d (%s) 耗时: %v\n", res.Source.Version, filepath.Base(res.Source.Path), res.Duration)
	return nil
}

// runMigrateUpTo 迁移到指定版本
func runMigrateUpTo(cmd *cobra.Command, args []string) error {
	targetVersion, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || targetVersion < 0 {
		return fmt.Errorf("无效的版本号：%q（版本号必须为非负整数）", args[0])
	}

	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 60*time.Second)
	defer cancel()

	fmt.Fprintf(cmd.OutOrStdout(), "正在迁移至版本 %d ...\n", targetVersion)
	results, err := provider.UpTo(ctx, targetVersion)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "已达到指定版本或更高版本，无需迁移")
			return nil
		}
		return fmt.Errorf("迁移到指定版本失败：%w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "已是指定版本，没有新变更应用")
		return nil
	}

	for _, res := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "-> 已应用版本 %d (%s) 耗时: %v\n", res.Source.Version, filepath.Base(res.Source.Path), res.Duration)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "迁移执行完成")
	return nil
}

// runMigrateDown 回滚最后一个迁移
func runMigrateDown(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 30*time.Second)
	defer cancel()

	fmt.Fprintln(cmd.OutOrStdout(), "正在回滚最后一个迁移...")
	res, err := provider.Down(ctx)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "没有可回滚的迁移")
			return nil
		}
		return fmt.Errorf("回滚迁移失败：%w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "-> 已回滚版本 %d (%s) 耗时: %v\n", res.Source.Version, filepath.Base(res.Source.Path), res.Duration)
	fmt.Fprintln(cmd.OutOrStdout(), "回滚完成")
	return nil
}

// runMigrateDownTo 回滚到指定版本
func runMigrateDownTo(cmd *cobra.Command, args []string) error {
	targetVersion, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || targetVersion < 0 {
		return fmt.Errorf("无效的版本号：%q（版本号必须为非负整数）", args[0])
	}

	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 60*time.Second)
	defer cancel()

	fmt.Fprintf(cmd.OutOrStdout(), "正在回滚至版本 %d ...\n", targetVersion)
	results, err := provider.DownTo(ctx, targetVersion)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "已处于目标版本或更低版本，无需回滚")
			return nil
		}
		return fmt.Errorf("回滚到指定版本失败：%w", err)
	}

	for _, res := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "-> 已回滚版本 %d (%s) 耗时: %v\n", res.Source.Version, filepath.Base(res.Source.Path), res.Duration)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "回滚完成")
	return nil
}

// runMigrateRedo 回滚上一个迁移并重新执行
func runMigrateRedo(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 60*time.Second)
	defer cancel()

	fmt.Fprintln(cmd.OutOrStdout(), "正在执行重做（Redo: Down + Up）...")

	// 1. 回滚最后一个版本
	downRes, err := provider.Down(ctx)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "当前没有任何已应用的迁移可供重做")
			return nil
		}
		return fmt.Errorf("Redo 步骤 1 (回滚) 失败：%w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "-> 已回滚版本 %d (%s)\n", downRes.Source.Version, filepath.Base(downRes.Source.Path))
	// 2. 重新应用该版本
	upRes, err := provider.UpByOne(ctx)
	if err != nil {
		return fmt.Errorf("Redo 步骤 2 (重新应用) 失败：%w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "-> 已重新应用版本 %d (%s)\n", upRes.Source.Version, filepath.Base(upRes.Source.Path))

	fmt.Fprintln(cmd.OutOrStdout(), "Redo 操作完成")
	return nil
}

// runMigrateReset 回滚所有迁移
func runMigrateReset(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := getProvider(cfg, db)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(commandContext(cmd), 60*time.Second)
	defer cancel()

	fmt.Fprintln(cmd.OutOrStdout(), "正在回滚所有迁移至初始状态（Reset）...")
	results, err := provider.DownTo(ctx, 0)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			fmt.Fprintln(cmd.OutOrStdout(), "当前没有任何已应用的迁移")
			return nil
		}
		return fmt.Errorf("重置迁移失败：%w", err)
	}

	for _, res := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "-> 已回滚版本 %d (%s)\n", res.Source.Version, filepath.Base(res.Source.Path))
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Reset 重置完成")
	return nil
}

// runMigrateVersion 查看当前版本
func runMigrateVersion(cmd *cobra.Command, args []string) error {
	cfg, err := loadDBConfig(migrateFlags.config)
	if err != nil {
		return err
	}

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(commandContext(cmd), 10*time.Second)
	defer cancel()

	ver, err := migrate.GetDBVersion(ctx, db, cfg.Driver, migrate.WithTableName(migrateFlags.tableName))
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "当前数据库迁移版本: %d\n", ver)
	return nil
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// 全局 Persistent Flags
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.dir, "dir", "d", "migrations", "迁移文件所在目录")
	migrateCmd.PersistentFlags().StringVarP(&migrateFlags.config, "config", "c", "config/app.yaml", "应用配置文件路径")
	migrateCmd.PersistentFlags().StringVar(&migrateFlags.tableName, "table", goose.DefaultTablename, "迁移版本记录表名")

	// 注册子命令
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateUpByOneCmd)
	migrateCmd.AddCommand(migrateUpToCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateDownToCmd)
	migrateCmd.AddCommand(migrateRedoCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
}
