package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host          string
	Port          string
	User          string
	Password      string
	DBName        string
	SSLMode       string
	MaxOpenDbConn int
	MaxIdleDbConn time.Duration
	MaxDbLifeTime time.Duration
}

type Database interface {
	GetGormDB() *gorm.DB
	WithTx(tx *gorm.DB) Database
	GetDBConfig() Config
}

type PostgresDatabase struct {
	DB     *gorm.DB
	Config Config
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func NewPostgresDatabase(ctx context.Context, cfg *Config) (Database, error) {
	dsn := cfg.DSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenDbConn)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleDbConn)
	sqlDB.SetConnMaxLifetime(cfg.MaxDbLifeTime)

	return &PostgresDatabase{DB: db, Config: *cfg}, nil
}

func (p *PostgresDatabase) GetGormDB() *gorm.DB {
	return p.DB
}

func (p *PostgresDatabase) WithTx(tx *gorm.DB) Database {
	return &PostgresDatabase{DB: tx, Config: p.Config}
}

func (p *PostgresDatabase) GetDBConfig() Config {
	return p.Config
}
