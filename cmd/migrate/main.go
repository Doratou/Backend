package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/Doratou/Backend/bootstrap"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationsDir = "migrations"

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Usage: go run ./cmd/migrate [up|down|status|reset]")
	}
	command := args[0]

	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	dbCfg := cfg.Environment.DataBase
	dsn := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbCfg.User, dbCfg.Password),
		Host:     net.JoinHostPort(dbCfg.Host, dbCfg.Port),
		Path:     "/" + dbCfg.Name,
		RawQuery: url.Values{"sslmode": {dbCfg.SSLMode}}.Encode(),
	}

	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	ctx := context.Background()
	switch command {
	case "up":
		if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("goose up failed: %v", err)
		}
		fmt.Println("migrations applied successfully")
	case "down":
		if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("goose down failed: %v", err)
		}
		fmt.Println("migration rolled back successfully")
	case "status":
		if err := goose.StatusContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("goose status failed: %v", err)
		}
	case "reset":
		if err := goose.ResetContext(ctx, db, migrationsDir); err != nil {
			log.Fatalf("goose reset failed: %v", err)
		}
		fmt.Println("migrations reset successfully")
	default:
		log.Fatalf("unknown migration command %q", command)
	}
	os.Exit(0)
}
