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
	defer db.Close()

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
}
