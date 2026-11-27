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
	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/service"
	"github.com/Doratou/Backend/internal/application/service/decorator"
	"github.com/Doratou/Backend/internal/infrastructure/ai"
	brokerInfra "github.com/Doratou/Backend/internal/infrastructure/broker/rabbitmq"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	emailInfra "github.com/Doratou/Backend/internal/infrastructure/email"
	"github.com/Doratou/Backend/internal/infrastructure/jwt"
	oauthInfra "github.com/Doratou/Backend/internal/infrastructure/oauth"
	randomInfra "github.com/Doratou/Backend/internal/infrastructure/random"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	postgresRepo "github.com/Doratou/Backend/internal/infrastructure/repository/postgres"
	redisRepo "github.com/Doratou/Backend/internal/infrastructure/repository/redis"
	"github.com/Doratou/Backend/internal/infrastructure/seed"
	"github.com/Doratou/Backend/internal/infrastructure/storage"
	weatherInfra "github.com/Doratou/Backend/internal/infrastructure/weather"
	wsInfra "github.com/Doratou/Backend/internal/infrastructure/websocket"
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/Doratou/Backend/internal/presentation/v1/router"
	"github.com/Doratou/Backend/pkg/captcha"
	"github.com/Doratou/Backend/pkg/hasher"
	mail "github.com/Doratou/Backend/pkg/wneessen-go-mail"
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

	// 3. RustFS / MinIO S3 Object Storage connection
	rustfsEnv := cfg.Environment.RustFS
	rustfsCfg := &storage.Config{
		Endpoint:        rustfsEnv.Endpoint,
		AccessKeyID:     rustfsEnv.AccessKeyID,
		SecretAccessKey: rustfsEnv.SecretAccessKey,
		Bucket:          rustfsEnv.BucketName,
		UseSSL:          rustfsEnv.UseSSL,
		Region:          rustfsEnv.Region,
		MaxConcurrency:  rustfsEnv.MaxConcurrency,
	}
	rustfsConnCtx, cancelRustFS := context.WithTimeout(context.Background(), 10*time.Second)
	rustfsStorage, err := storage.NewRustFSRepository(rustfsConnCtx, rustfsCfg)
	cancelRustFS()
	if err != nil {
		log.Fatalf("failed to connect to RustFS storage: %v", err)
	}
	log.Print("RustFS storage connected and bucket verified")

	// 3.1 RabbitMQ Broker Connection & Publisher
	rmqEnv := cfg.Environment.RabbitMQ
	rmqCfg := brokerInfra.DefaultConfig(rmqEnv.Host, rmqEnv.Port, rmqEnv.User, rmqEnv.Password, rmqEnv.VHost)
	rmqCfg.ConnectionTimeout = cfg.Constant.RabbitMQ.DefaultPublishTimeout
	rmqCfg.InitialReconnectDelay = cfg.Constant.RabbitMQ.DefaultReconnectDelay

	rmqConnCtx, cancelRmq := context.WithTimeout(context.Background(), 10*time.Second)
	rmqConnMgr, err := brokerInfra.NewConnectionManager(rmqConnCtx, rmqCfg)
	cancelRmq()
	if err != nil {
		log.Printf("[RABBITMQ WARNING] failed to connect to RabbitMQ broker at startup: %v (broker will auto-reconnect)", err)
	} else {
		defer rmqConnMgr.Close()
		log.Print("RabbitMQ broker connected")

		// Setup Doratou notification exchange and queues with default safety net TTL
		setupCtx, cancelSetup := context.WithTimeout(context.Background(), 10*time.Second)
		topology := brokerInfra.NewTopology(rmqConnMgr)
		qTTL := cfg.Environment.QueueTTL
		if err := topology.SetupDoratouTopology(setupCtx, qTTL.EmailQueueTTL, qTTL.SMSQueueTTL, qTTL.DLQMessageTTL); err != nil {
			log.Printf("[RABBITMQ WARNING] failed to declare doratou topology: %v", err)
		} else {
			log.Print("RabbitMQ Doratou notification topology initialized")
		}
		cancelSetup()
	}
	rmqPublisher := brokerInfra.NewPublisher(rmqConnMgr)

	// 4. Infrastructure Transaction Managers
	dbTrxManager := database.NewTrxManager(postgresDB)
	redisTrxManager := redisInfra.NewTrxManager(redisCache, 3, 20*time.Millisecond)

	// 5. Repositories
	userRepo := postgresRepo.NewUserRepository(postgresDB)
	attachmentRepo := postgresRepo.NewAttachmentRepository(postgresDB)
	adminReportRepo := postgresRepo.NewAdminReportRepository(postgresDB)
	friendshipRepo := postgresRepo.NewFriendshipRepository(postgresDB)
	groupRepo := postgresRepo.NewGroupRepository(postgresDB)
	pollRepo := postgresRepo.NewPollRepository(postgresDB)
	checklistRepo := postgresRepo.NewChecklistRepository(postgresDB)
	expenseRepo := postgresRepo.NewExpenseRepository(postgresDB)
	settlementRepo := postgresRepo.NewSettlementRepository(postgresDB)
	financialReportRepo := postgresRepo.NewFinancialReportRepository(postgresDB)
	luckyDrawRepo := postgresRepo.NewLuckyDrawRepository(postgresDB)
	notificationRepo := postgresRepo.NewNotificationRepository(postgresDB)
	tokenPriceRepo := postgresRepo.NewTokenPriceRepository(postgresDB)
	walletRepo := postgresRepo.NewWalletRepository(postgresDB)
	tripSuggestionRepo := postgresRepo.NewTripSuggestionRepository(postgresDB)
	inviteRepo := postgresRepo.NewInviteRepository(postgresDB)
	oauthRepo := postgresRepo.NewOAuthRepository(postgresDB)
	otpRepo := redisRepo.NewOTPRepository(redisCache)
	tokenRepo := redisRepo.NewAuthnTokenRepository(redisCache)
	oauthStateRepo := redisRepo.NewOAuthStateRepository(redisCache)
	captchaRepo := redisRepo.NewCaptchaRepository(redisCache)

	// 6. Crypto & Security Utilities
	passwordHasher := hasher.NewBcryptHasher()
	captchaMan := captcha.NewDigitCaptcha(nil)
	jwtMaker, err := jwt.NewJWTMaker(cfg.Environment.JWT.SecretKey)
	if err != nil {
		log.Fatalf("failed to initialize JWT maker: %v", err)
	}

	// 6.1 Database Seeder (idempotent: seeds regular user & admin user if not present)
	seedCtx, cancelSeed := context.WithTimeout(context.Background(), 15*time.Second)
	if err := seed.Run(seedCtx, userRepo, postgresDB, groupRepo, inviteRepo, passwordHasher, cfg.Environment.Seed); err != nil {
		log.Printf("[SEED WARNING] database seeder finished with notice/error: %v", err)
	}
	cancelSeed()

	// 6.2 Google OAuth Provider
	var googleProvider oauthInfra.GoogleProvider
	if cfg.Environment.Google.ClientID != "" {
		googleCtx, cancelGoogle := context.WithTimeout(context.Background(), 10*time.Second)
		gp, err := oauthInfra.NewGoogleProvider(googleCtx, oauthInfra.GoogleConfig{
			ClientID:     cfg.Environment.Google.ClientID,
			ClientSecret: cfg.Environment.Google.ClientSecret,
			RedirectURL:  cfg.Environment.OAuth.Protcol + "://" + cfg.Environment.OAuth.Domain + ":" + cfg.Environment.OAuth.Port + cfg.Environment.Google.RedirectEndPoint,
		})
		cancelGoogle()
		if err != nil {
			log.Printf("[GOOGLE OAUTH WARNING] failed to initialize Google OAuth provider: %v", err)
		} else {
			googleProvider = gp
			log.Print("Google OAuth provider initialized")
		}
	} else {
		log.Print("[GOOGLE OAUTH NOTICE] GOOGLE_CLIENT_ID is not configured; Google OAuth login is disabled")
	}

	// 7. Application Services
	emailSender := mail.NewMailer(mail.EmailConfig{
		From:     cfg.Environment.Email.DoratouEmail,
		Password: cfg.Environment.Email.DoratouEmailAppPassword,
		SMTPHost: cfg.Environment.Email.SMTPHost,
		SMTPPort: cfg.Environment.Email.SMTPPort,
	})
	emailRenderer, err := emailInfra.NewOTPEmailRenderer(cfg.Environment.Email.FontBaseURL)
	if err != nil {
		log.Fatalf("failed to initialize OTP email template: %v", err)
	}
	emailFontFiles, err := emailInfra.OTPEmailFontFiles()
	if err != nil {
		log.Fatalf("failed to initialize OTP email fonts: %v", err)
	}
	otpService := service.NewOTPService(
		otpRepo,
		redisTrxManager,
		emailSender,
		emailRenderer,
		rmqPublisher,
		service.OTPServiceConfig{
			TTL:                  cfg.Environment.OTP.TTL,
			VerifiedTTL:          cfg.Environment.OTP.VerifiedTTL,
			Digits:               cfg.Environment.OTP.Digits,
			EmailDeliveryEnabled: cfg.Environment.Email.DeliveryEnabled,
			EmailDeliveryMode:    cfg.Environment.Email.DeliveryMode,
		},
		cfg.Constant,
	)

	authService := service.NewAuthService(
		userRepo,
		oauthRepo,
		attachmentRepo,
		otpService,
		tokenRepo,
		oauthStateRepo,
		googleProvider,
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

	captchaService := service.NewCaptchaService(
		captchaRepo,
		captchaMan,
		service.CaptchaServiceConfig{
			TTL: cfg.Environment.Captcha.TTL,
		},
		cfg.Constant,
	)
	var groupService adapter.GroupService = service.NewGroupService(groupRepo, attachmentRepo)
	var inviteService adapter.InviteService = service.NewInviteService(groupRepo, inviteRepo, friendshipRepo, userRepo, passwordHasher, cfg.Environment.Invite.SecretKey, cfg.Environment.Invite.DirectInviteRejectCooldown)
	pollService := service.NewPollService(pollRepo, userRepo, dbTrxManager)
	luckyDrawService := service.NewLuckyDrawService(luckyDrawRepo, userRepo, dbTrxManager, randomInfra.NewCryptoGenerator())

	adminReportService := service.NewAdminReportService(
		adminReportRepo,
		userRepo,
	)
	tokenPriceService := service.NewTokenPriceService(tokenPriceRepo)
	weatherService := service.NewWeatherService(weatherInfra.NewCachedWeatherProvider(weatherInfra.NewOpenMeteoClient(), redisCache))
	aiModel := cfg.Environment.AvalAI.Model
	tripSuggestionService := service.NewTripSuggestionService(ai.NewAvalAIClient(ai.AvalAIConfig{
		APIKey: cfg.Environment.AvalAI.APIKey, BaseURL: cfg.Environment.AvalAI.BaseURL, Model: aiModel,
	}), weatherService, tripSuggestionRepo, aiModel)

	attachmentService := service.NewAttachmentService(
		attachmentRepo,
		groupRepo,
		expenseRepo,
		settlementRepo,
		userRepo,
		dbTrxManager,
		rustfsStorage,
		service.AttachmentServiceConfig{
			PublicURLPrefix:     rustfsEnv.PublicURLPrefix,
			DefaultPresignedTTL: cfg.Constant.Attachment.DefaultPresignedTTL,
		},
		cfg.Constant,
	)
	settlementService := service.NewSettlementService(settlementRepo, financialReportRepo, userRepo, dbTrxManager, attachmentService)
	expenseService := service.NewExpenseService(expenseRepo, userRepo, dbTrxManager, attachmentService)
	readTrxManager := database.NewReadSnapshotTrxManager(postgresDB)
	walletCfg := cfg.Environment.Wallet
	walletService := service.NewWalletService(walletRepo, userRepo, tokenPriceRepo, dbTrxManager, readTrxManager, randomInfra.NewCryptoGenerator(), service.WalletServiceConfig{
		MockRandomEnabled: walletCfg.MockRandomEnabled, MockSuccessPercent: walletCfg.MockSuccessPercent,
	})
	financialReportService := service.NewFinancialReportService(financialReportRepo, userRepo, readTrxManager, attachmentService)
	checklistService := service.NewChecklistService(checklistRepo, dbTrxManager, readTrxManager, attachmentService)
	friendshipService := service.NewFriendshipService(
		friendshipRepo,
		userRepo,
		dbTrxManager,
		readTrxManager,
		attachmentService,
		service.FriendshipServiceConfig{
			RequestCooldown: cfg.Environment.Friendship.RequestCooldown,
		},
		cfg.Constant,
	)

	userService := service.NewUserService(
		userRepo,
		attachmentService,
		otpService,
		tokenRepo,
		jwtMaker,
		redisTrxManager,
	)

	var wsRuntime *wsInfra.Runtime
	if wsCfg := cfg.ResolveWS(); wsCfg.Enabled {
		wsRuntime = wsInfra.NewRuntime(wsInfra.Config{
			RequireAuth:        wsCfg.AuthRequired,
			InsecureSkipVerify: wsCfg.InsecureSkipVerify,
			OriginPatterns:     wsCfg.OriginPatterns,
			ReadLimit:          wsCfg.ReadLimitBytes,
			SendBufferSize:     wsCfg.SendBufferSize,
			WriteTimeout:       wsCfg.WriteTimeout,
			PingInterval:       wsCfg.PingInterval,
			PongTimeout:        wsCfg.PongTimeout,
			ShutdownTimeout:    wsCfg.ShutdownTimeout,
			MaxConnsPerUser:    wsCfg.MaxConnsPerUser,
			SlowConsumerPolicy: wsInfra.SlowConsumerPolicy(wsCfg.SlowConsumerPolicy),
			PubSubChannel:      wsCfg.PubSubChannel,
		}, redisCache.GetClient(), jwtMaker, tokenRepo)
	}

	notificationService := service.NewNotificationService(
		notificationRepo,
		groupRepo,
		dbTrxManager,
		wsRuntime.Publisher(),
		service.NotificationServiceConfig{
			RealtimeTTL: cfg.Environment.Notification.RealtimeTTL,
			StandardTTL: cfg.Environment.Notification.StandardTTL,
			NormalTTL:   cfg.Environment.Notification.NormalTTL,
		},
	)

	// Decorate domain services with real-time & persistent event notification dispatching.
	friendshipService = decorator.NewFriendshipNotificationDecorator(friendshipService, notificationService, friendshipRepo, userRepo)
	settlementService = decorator.NewSettlementNotificationDecorator(settlementService, notificationService, groupRepo)
	pollService = decorator.NewPollNotificationDecorator(pollService, notificationService, groupRepo, userRepo)
	luckyDrawService = decorator.NewLuckyDrawNotificationDecorator(luckyDrawService, notificationService, groupRepo, userRepo)
	checklistService = decorator.NewChecklistNotificationDecorator(checklistService, notificationService, groupRepo, userRepo, checklistRepo)
	adminReportService = decorator.NewAdminReportNotificationDecorator(adminReportService, notificationService, userRepo)
	groupService = decorator.NewGroupNotificationDecorator(groupService, notificationService, groupRepo, userRepo)
	inviteService = decorator.NewInviteNotificationDecorator(inviteService, notificationService, groupRepo, userRepo)
	tokenPriceService = decorator.NewTokenPriceNotificationDecorator(tokenPriceService, notificationService)

	// 7.1 Optional End-to-End Notification Demo Seeder (runs all 14 real event flows on startup)
	if cfg.Environment.Notification.SeedDemoEvents {
		seedNotifCtx, cancelSeedNotif := context.WithTimeout(context.Background(), 30*time.Second)
		if err := seed.SeedNotificationEvents(seedNotifCtx, seed.NotificationSeedDeps{
			DB:                  postgresDB,
			FriendshipService:   friendshipService,
			SettlementService:   settlementService,
			PollService:         pollService,
			LuckyDrawService:    luckyDrawService,
			ChecklistService:    checklistService,
			AdminReportService:  adminReportService,
			GroupService:        groupService,
			InviteService:       inviteService,
			TokenPriceService:   tokenPriceService,
			NotificationService: notificationService,
		}); err != nil {
			log.Printf("[SEED NOTIFICATIONS WARNING] failed to emit demo notification events: %v", err)
		}
		cancelSeedNotif()
	}

	// 8. Presentation Handlers & Routers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(&cfg.Environment.OAuth, authService)
	captchaHandler := handler.NewCaptchaHandler(captchaService)

	attachmentHandler := handler.NewAttachmentHandler(attachmentService)

	adminReportHandler := handler.NewAdminReportHandler(adminReportService)
	hamyarHandler := handler.NewHamyarHandler(tripSuggestionService, weatherService, groupService)

	friendshipHandler := handler.NewFriendshipHandler(friendshipService)
	groupHandler := handler.NewGroupHandler(groupService)
	pollHandler := handler.NewPollHandler(pollService)
	checklistHandler := handler.NewChecklistHandler(checklistService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	settlementHandler := handler.NewSettlementHandler(settlementService)
	financialReportHandler := handler.NewFinancialReportHandler(financialReportService)
	luckyDrawHandler := handler.NewLuckyDrawHandler(luckyDrawService)
	inviteHandler := handler.NewInviteHandler(inviteService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	tokenPriceHandler := handler.NewTokenPriceHandler(tokenPriceService)
	walletHandler := handler.NewWalletHandler(walletService)
	var wsHandler *handler.WSHandler
	if wsRuntime != nil {
		wsHandler = handler.NewWSHandler(wsRuntime.HTTPHandler(), wsRuntime.Monitor())
	}

	engine := router.New(router.Dependencies{
		AuthHandler:            authHandler,
		UserHandler:            userHandler,
		CaptchaHandler:         captchaHandler,
		AttachmentHandler:      attachmentHandler,
		AdminReportHandler:     adminReportHandler,
		NotificationHandler:    notificationHandler,
		TokenPriceHandler:      tokenPriceHandler,
		WalletHandler:          walletHandler,
		WSHandler:              wsHandler,
		GroupHandler:           groupHandler,
		InviteHandler:          inviteHandler,
		FriendshipHandler:      friendshipHandler,
		PollHandler:            pollHandler,
		ChecklistHandler:       checklistHandler,
		ExpenseHandler:         expenseHandler,
		SettlementHandler:      settlementHandler,
		FinancialReportHandler: financialReportHandler,
		LuckyDrawHandler:       luckyDrawHandler,
		HamyarHandler:          hamyarHandler,
		CaptchaService:         captchaService,
		CaptchaEnabled:         cfg.Environment.Captcha.Enabled,
		TokenMaker:             jwtMaker,
		TokenRepo:              tokenRepo,
		EmailFontFiles:         emailFontFiles,
	})

	// 9. HTTP Server Lifecycle & Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wsRuntime.Start(ctx)

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
	wsRuntime.Shutdown()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during HTTP server shutdown: %v", err)
	}
	log.Print("Server stopped")
}
