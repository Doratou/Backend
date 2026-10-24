package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/application/service"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	postgresRepo "github.com/Doratou/Backend/internal/infrastructure/repository/postgres"
	redisRepo "github.com/Doratou/Backend/internal/infrastructure/repository/redis"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/Doratou/Backend/internal/presentation/v1/router"
	"github.com/Doratou/Backend/pkg/hasher"
)

func main() {
	cfg, err := bootstrap.Run()
	if err != nil {
		log.Fatalf("bootstrap error: %v", err)
	}

	// 1. PostgreSQL connection with GORM & connection pool
	dbEnv := cfg.Environment.DataBase
	dbCfg := &database.Config{
		Host:          dbEnv.Host,
		Port:          dbEnv.Port,
		User:          dbEnv.User,
		Password:      dbEnv.Password,
		DBName:        dbEnv.Name,
		SSLMode:       dbEnv.SSLMode,
		MaxOpenDbConn: int(cfg.Constant.Database.MaxOpenDbConn),
		MaxIdleDbConn: cfg.Constant.Database.MaxIdleDbConn,
		MaxDbLifeTime: cfg.Constant.Database.MaxDbLifeTime,
	}
	dbConnCtx, cancelDb := context.WithTimeout(context.Background(), 10*time.Second)
	postgresDB, err := database.NewPostgresDatabase(dbConnCtx, dbCfg)
	cancelDb()
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	sqlDB, err := postgresDB.GetGormDB().DB()
	if err == nil {
		defer sqlDB.Close()
	}
	log.Print("PostgreSQL connected (GORM)")

	// 2. Redis connection
	redisEnv := cfg.Environment.Redis
	redisDBNum, err := strconv.Atoi(redisEnv.RDBNumber)
	if err != nil || redisDBNum < 0 {
		log.Fatal("RDB_DB_NUMBER must be a non-negative integer")
	}
	redisCfg := &redisInfra.Config{
		Host:      redisEnv.Host,
		Port:      redisEnv.Port,
		Password:  redisEnv.Password,
		RDBNumber: redisDBNum,
	}
	redisConnCtx, cancelRedis := context.WithTimeout(context.Background(), 10*time.Second)
	redisCache, err := redisInfra.NewRedisDatabase(redisConnCtx, redisCfg)
	cancelRedis()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisCache.GetClient().Close()
	log.Print("Redis connected")

	// 3. Infrastructure Transaction Managers
	dbTrxManager := database.NewTrxManager(postgresDB)
	redisTrxManager := redisInfra.NewTrxManager(redisCache, 3, 20*time.Millisecond)

	// 4. Repositories
	userRepo := postgresRepo.NewUserRepository(postgresDB)
	otpRepo := redisRepo.NewOTPRepository(redisCache)
	tokenRepo := redisRepo.NewAuthnTokenRepository(redisCache)

	// 5. Crypto & Security Utilities
	passwordHasher := hasher.NewBcryptHasher()
	jwtMaker, err := jwt.NewJWTMaker(cfg.Environment.JWT.SecretKey)
	if err != nil {
		log.Fatalf("failed to initialize JWT maker: %v", err)
	}

	// 6. Application Services
	otpService := service.NewOTPService(
		otpRepo,
		redisTrxManager,
		service.OTPServiceConfig{
			TTL:         cfg.Environment.OTP.TTL,
			VerifiedTTL: cfg.Environment.OTP.VerifiedTTL,
			Digits:      cfg.Environment.OTP.Digits,
		},
		cfg.Constant,
	)

	authService := service.NewAuthService(
		userRepo,
		otpService,
		tokenRepo,
		dbTrxManager,
		redisTrxManager,
		jwtMaker,
		passwordHasher,
		service.AuthServiceConfig{
			AccessTokenTTL:  cfg.Environment.JWT.AccessTokenTTL,
			RefreshTokenTTL: cfg.Environment.JWT.RefreshTokenTTL,
		},
		cfg.Constant,
	)

	// 7. Presentation Handlers & Routers
	authHandler := handler.NewAuthHandler(authService)
	engine := router.New(router.Dependencies{
		AuthHandler: authHandler,
		TokenMaker:  jwtMaker,
		TokenRepo:   tokenRepo,
	})

	// 8. HTTP Server Lifecycle & Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:              net.JoinHostPort(cfg.Environment.HttpServer.Domain, cfg.Environment.HttpServer.Port),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("Shutdown signal received, shutting down gracefully...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during HTTP server shutdown: %v", err)
	}
	log.Print("Server stopped")
}
