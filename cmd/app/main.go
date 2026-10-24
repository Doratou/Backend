package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/presentation/v1/router"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatal(err)
	}
	db := cfg.Environment.DataBase
	dsn := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.User, db.Password),
		Host:     net.JoinHostPort(db.Host, db.Port),
		Path:     "/" + db.Name,
		RawQuery: url.Values{"sslmode": {db.SSLMode}}.Encode(),
	}
	dbConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		log.Fatal("invalid PostgreSQL connection settings")
	}
	dbConfig.MaxConns = cfg.Constant.Database.MaxOpenDbConn
	dbConfig.MaxConnIdleTime = cfg.Constant.Database.MaxIdleDbConn
	dbConfig.MaxConnLifetime = cfg.Constant.Database.MaxDbLifeTime
	dbConfig.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = pool.Ping(pingCtx)
	cancel()
	if err != nil {
		log.Fatal("PostgreSQL connection failed: ", err)
	}
	log.Print("PostgreSQL connected")

	redisConfig := cfg.Environment.Redis
	redisDB, err := strconv.Atoi(redisConfig.RDBNumber)
	if err != nil || redisDB < 0 {
		log.Fatal("RDB_DB_NUMBER must be a non-negative integer")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:                  net.JoinHostPort(redisConfig.Host, redisConfig.Port),
		Password:              redisConfig.Password,
		DB:                    redisDB,
		ContextTimeoutEnabled: true,
	})
	defer redisClient.Close()
	pingCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	err = redisClient.Ping(pingCtx).Err()
	cancel()
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	log.Print("Redis connected")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{
		Addr:              net.JoinHostPort(cfg.Environment.HttpServer.Domain, cfg.Environment.HttpServer.Port),
		Handler:           router.New(router.Dependencies{}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Print(err)
	}
}
