// Package migrate provides lightweight database migration helpers based on pressly/goose/v3.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

// Option defines a functional option for configuring migration executions.
type Option func(*options)

type options struct {
	tableName string
}

// WithTableName configures a custom table name for tracking migration versions (default: "goose_db_version").
func WithTableName(table string) Option {
	return func(o *options) {
		if strings.TrimSpace(table) != "" {
			o.tableName = strings.TrimSpace(table)
		}
	}
}

// ToGooseDialect maps framework database driver names (mysql, postgres, sqlite) to standard goose.Dialect.
func ToGooseDialect(driver string) (goose.Dialect, error) {
	d := strings.ToLower(strings.TrimSpace(driver))
	switch d {
	case "mysql":
		return goose.DialectMySQL, nil
	case "postgres", "postgresql", "pgsql", "pgx":
		return goose.DialectPostgres, nil
	case "sqlite", "sqlite3":
		return goose.DialectSQLite3, nil
	case "clickhouse":
		return goose.DialectClickHouse, nil
	case "mssql", "sqlserver":
		return goose.DialectMSSQL, nil
	default:
		return "", fmt.Errorf("unsupported database driver for migrate: %q (supported: mysql, postgres, sqlite3)", driver)
	}
}

// resolveFS prepares the target fs.FS. If dir is not empty and not ".", it attempts fs.Sub.
func resolveFS(fsys fs.FS, dir string) (fs.FS, error) {
	if fsys == nil {
		return nil, fmt.Errorf("migration filesystem cannot be nil")
	}
	// Normalize Windows backslashes and relative prefixes
	cleanDir := filepath.ToSlash(strings.TrimSpace(dir))
	cleanDir = strings.TrimPrefix(cleanDir, "./")
	cleanDir = strings.Trim(cleanDir, "/")
	if cleanDir == "" || cleanDir == "." {
		return fsys, nil
	}
	sub, err := fs.Sub(fsys, cleanDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open sub migration directory %q: %w", cleanDir, err)
	}
	return sub, nil
}

// NewProvider creates a new goose.Provider instance with isolated state.
func NewProvider(db *sql.DB, driver string, fsys fs.FS, dir string, opts ...Option) (*goose.Provider, error) {
	dialect, err := ToGooseDialect(driver)
	if err != nil {
		return nil, err
	}

	subFS, err := resolveFS(fsys, dir)
	if err != nil {
		return nil, err
	}

	cfg := &options{
		tableName: goose.DefaultTablename,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	var providerOpts []goose.ProviderOption
	if cfg.tableName != "" && cfg.tableName != goose.DefaultTablename {
		providerOpts = append(providerOpts, goose.WithTableName(cfg.tableName))
	}

	provider, err := goose.NewProvider(dialect, db, subFS, providerOpts...)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// RunUp executes all pending migrations from a local filesystem directory.
func RunUp(ctx context.Context, db *sql.DB, driver string, dir string, opts ...Option) error {
	cleanDir := strings.TrimSpace(dir)
	if cleanDir == "" {
		cleanDir = "migrations"
	}
	return RunUpFS(ctx, db, driver, os.DirFS(cleanDir), ".", opts...)
}

// RunUpFS executes all pending migrations from a given fs.FS (e.g. embed.FS or os.DirFS).
func RunUpFS(ctx context.Context, db *sql.DB, driver string, fsys fs.FS, dir string, opts ...Option) error {
	provider, err := NewProvider(db, driver, fsys, dir, opts...)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// RunDown rolls back the most recently applied migration from a local filesystem directory.
func RunDown(ctx context.Context, db *sql.DB, driver string, dir string, opts ...Option) error {
	cleanDir := strings.TrimSpace(dir)
	if cleanDir == "" {
		cleanDir = "migrations"
	}
	return RunDownFS(ctx, db, driver, os.DirFS(cleanDir), ".", opts...)
}

// RunDown rolls back the most recently applied migration from a given fs.FS.
func RunDownFS(ctx context.Context, db *sql.DB, driver string, fsys fs.FS, dir string, opts ...Option) error {
	provider, err := NewProvider(db, driver, fsys, dir, opts...)
	if err != nil {
		return err
	}
	_, err = provider.Down(ctx)
	if err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}

// GetStatus returns the current migration status list.
func GetStatus(ctx context.Context, db *sql.DB, driver string, fsys fs.FS, dir string, opts ...Option) ([]*goose.MigrationStatus, error) {
	provider, err := NewProvider(db, driver, fsys, dir, opts...)
	if err != nil {
		return nil, err
	}
	status, err := provider.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}
	return status, nil
}

// GetDBVersion returns the current database version.
func GetDBVersion(ctx context.Context, db *sql.DB, driver string, opts ...Option) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database connection cannot be nil")
	}

	dialect, err := ToGooseDialect(driver)
	if err != nil {
		return 0, err
	}

	cfg := &options{
		tableName: goose.DefaultTablename,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	store, err := database.NewStore(dialect, cfg.tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to create database store: %w", err)
	}

	ver, err := store.GetLatestVersion(ctx, db)
	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		if errors.Is(err, database.ErrVersionNotFound) {
			return 0, nil
		}
		// When the version table does not exist yet, treat it as version 0.
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "no such table") || strings.Contains(errStr, "doesn't exist") {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to query database migration version: %w", err)
	}
	return ver, nil
}
