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
	"todoapp/internal/controller/restapi/middleware"
	"todoapp/internal/repository/postgres"
	"todoapp/internal/usecases/task"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func Run(cfg *config.Config) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger.Info("starting application")

	ctx := context.Background()
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.PostgresDSN)
	if err != nil {
		logger.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("closing database connection failed", slog.String("error", err.Error()))
		}
	}(db)

	if err := db.PingContext(ctx); err != nil {
		logger.Error("database ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database connected successfully")

	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepo(db)

	authSvc := middleware.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpire)
	taskSvc := task.NewService(taskRepo)

	authHandler := restapi.NewAuthHandler(authSvc)
	taskHandler := restapi.NewTaskHandler(taskSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/tasks", taskHandler.Create)
	mux.HandleFunc("GET /api/tasks", taskHandler.GetByUser)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tasksMux := http.NewServeMux()
	tasksMux.HandleFunc("GET /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tasks":[]}`))
	})

	mux.Handle("/api/tasks", middleware.AuthMiddleware(cfg.JWTSecret, tasksMux))

	srv := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server started", slog.String("addr", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
