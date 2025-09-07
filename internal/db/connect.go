package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"github.com/pressly/goose/v3"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Type     string `json:"type"`     // sqlite, postgres, mysql
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	SSLMode  string `json:"ssl_mode,omitempty"`
	DataDir  string `json:"data_dir,omitempty"` // For SQLite
}

// Connect connects to the database based on the configuration
func Connect(ctx context.Context, config *DatabaseConfig) (*sql.DB, error) {
	switch strings.ToLower(config.Type) {
	case "sqlite", "":
		return connectSQLite(ctx, config)
	case "postgres", "postgresql":
		return connectPostgres(ctx, config)
	case "mysql":
		return connectMySQL(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// connectSQLite connects to SQLite database
func connectSQLite(ctx context.Context, config *DatabaseConfig) (*sql.DB, error) {
	dataDir := config.DataDir
	if dataDir == "" {
		return nil, fmt.Errorf("data.dir is not set for SQLite")
	}
	
	dbPath := config.Database
	if dbPath == "" {
		dbPath = "crush.db"
	}
	
	// If not absolute path, make it relative to dataDir
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(dataDir, dbPath)
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set pragmas for better performance
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA page_size = 4096;",
		"PRAGMA cache_size = -8000;",
		"PRAGMA synchronous = NORMAL;",
	}

	for _, pragma := range pragmas {
		if _, err = db.ExecContext(ctx, pragma); err != nil {
			slog.Error("Failed to set pragma", pragma, err)
		} else {
			slog.Debug("Set pragma", "pragma", pragma)
		}
	}

	return applyMigrations(db, "sqlite3")
}

// connectPostgres connects to PostgreSQL database
func connectPostgres(ctx context.Context, config *DatabaseConfig) (*sql.DB, error) {
	host := config.Host
	if host == "" {
		host = "localhost"
	}
	
	port := config.Port
	if port == 0 {
		port = 5432
	}
	
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}
	
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, config.Username, config.Password, config.Database, sslMode)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}
	
	// Verify connection
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}
	
	return applyMigrations(db, "postgres")
}

// connectMySQL connects to MySQL database
func connectMySQL(ctx context.Context, config *DatabaseConfig) (*sql.DB, error) {
	host := config.Host
	if host == "" {
		host = "localhost"
	}
	
	port := config.Port
	if port == 0 {
		port = 3306
	}
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.Username, config.Password, host, port, config.Database)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %w", err)
	}
	
	// Verify connection
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
	}
	
	return applyMigrations(db, "mysql")
}

// applyMigrations applies database migrations
func applyMigrations(db *sql.DB, dialect string) (*sql.DB, error) {
	goose.SetBaseFS(FS)

	if err := goose.SetDialect(dialect); err != nil {
		slog.Error("Failed to set dialect", "error", err)
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		slog.Error("Failed to apply migrations", "error", err)
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	
	return db, nil
}

// Legacy function for backward compatibility
func ConnectSQLite(ctx context.Context, dataDir string) (*sql.DB, error) {
	config := &DatabaseConfig{
		Type:     "sqlite",
		Database: "crush.db",
		DataDir:  dataDir,
	}
	return Connect(ctx, config)
}
