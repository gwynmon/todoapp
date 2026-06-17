package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todoapp/config"
	"todoapp/internal/controller/restapi"
	"todoapp/internal/repository/postgres"
	"todoapp/internal/usecases/auth"
	"todoapp/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func RunAuth(cfg *config.Config) {
	log := logger.Init(cfg.LogLevel)

	ctx := context.Background()

	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.PostgresDSN)
	if err != nil {
		log.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("database close failed",
				slog.String("error", err.Error()),
			)
		}
	}()

	userRepo := postgres.NewUserRepository(db)

	authSvc := auth.NewAuthService(
		userRepo,
		cfg.JWTSecret,
		cfg.JWTExpire,
	)

	authHandler := restapi.NewAuthHandler(authSvc, log)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)

	srv := &http.Server{
		Addr:    cfg.AuthServerPort,
		Handler: mux,
	}

	go func() {
		log.Info("auth service started")
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", slog.String("error", err.Error()))
		}
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-signalCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", slog.String("error", err.Error()))
	}
}
