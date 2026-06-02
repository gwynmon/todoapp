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
	"todoapp/internal/repository/mongo"
	"todoapp/internal/repository/postgres"
	"todoapp/internal/usecases/note"
	"todoapp/internal/usecases/task"
	"todoapp/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	mongoDriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Run(cfg *config.Config) {
	log := logger.Init("info")

	ctx := context.Background()
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.PostgresDSN)
	if err != nil {
		log.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Error("closing database connection failed", slog.String("error", err.Error()))
		}
	}(db)

	if err := db.PingContext(ctx); err != nil {
		log.Error("database ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("database connected successfully")

	clientOpts := options.Client().ApplyURI(cfg.MongoDSN)
	mongoClient, err := mongoDriver.Connect(clientOpts)
	if err != nil {
		log.Error("mongo connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		log.Error("mongo ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer mongoClient.Disconnect(ctx)

	noteRepo := mongo.NewNoteRepo(mongoClient.Database("tododb"))
	noteSvc := note.NewService(noteRepo)
	noteHandler := restapi.NewNoteHandler(noteSvc)

	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepo(db)

	authSvc := middleware.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpire)
	taskSvc := task.NewService(taskRepo, noteRepo)

	authHandler := restapi.NewAuthHandler(authSvc, log)
	healthHandler := restapi.NewHealthHandler(db, mongoClient, log)
	taskHandler := restapi.NewTaskHandler(taskSvc, log)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)

	mux.HandleFunc("GET /healthz", healthHandler.Liveness)
	mux.HandleFunc("GET /readyz", healthHandler.Readiness)

	protected := func(h http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(cfg.JWTSecret, h)
	}

	mux.Handle("POST /api/tasks", protected(taskHandler.Create))
	mux.Handle("GET /api/tasks", protected(taskHandler.GetByUser))
	mux.Handle("GET /api/tasks/{taskID}", protected(taskHandler.GetByID))
	mux.Handle("PUT /api/tasks/{taskID}", protected(taskHandler.Update))
	mux.Handle("DELETE /api/tasks/{taskID}", protected(taskHandler.Delete))
	mux.Handle("POST /api/tasks/{taskID}/notes", protected(noteHandler.Create))
	mux.Handle("GET /api/tasks/{taskID}/notes", protected(noteHandler.List))
	mux.Handle("DELETE /api/notes/{noteID}", protected(noteHandler.Delete))

	srv := &http.Server{
		Addr:         cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("http server started", slog.String("addr", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped")
}
