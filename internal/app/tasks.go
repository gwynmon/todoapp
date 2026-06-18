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
	"todoapp/internal/usecases/notification"
	"todoapp/internal/usecases/task"
	"todoapp/pkg/broker"
	"todoapp/pkg/cache"
	"todoapp/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	mongoDriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func RunTasks(cfg *config.Config) {
	log := logger.Init(cfg.LogLevel)

	ctx := context.Background()

	// Postgres
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.PostgresDSN)
	if err != nil {
		log.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("closing database connection failed", slog.String("error", err.Error()))
		}
	}()
	if err := db.PingContext(ctx); err != nil {
		log.Error("database ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("postgres connected successfully")

	// MongoDB
	mongoClient, err := mongoDriver.Connect(options.Client().ApplyURI(cfg.MongoDSN))
	if err != nil {
		log.Error("mongo connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := mongoClient.Ping(pingCtx, nil); err != nil {
		log.Error("mongo ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	cancel()
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Error("mongo disconnect failed", slog.String("error", err.Error()))
		}
	}()
	log.Info("mongodb connected successfully")

	// Redis
	redisOpts, err := redis.ParseURL(cfg.RedisDSN)
	if err != nil {
		log.Error("redis dsn parse failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOpts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("redis connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Error("redis close failed", slog.String("error", err.Error()))
		}
	}()
	log.Info("redis connected successfully")

	// RabbitMQ
	rabbitConn, err := amqp.Dial(cfg.RabbitMQDSN)
	if err != nil {
		log.Error("rabbitmq connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := rabbitConn.Close(); err != nil {
			log.Error("rabbitmq close failed", slog.String("error", err.Error()))
		}
	}()
	producer, err := broker.NewProducer(rabbitConn)
	if err != nil {
		log.Error("rabbitmq producer failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer producer.Close()
	log.Info("rabbitmq connected successfully")

	// Repositories & Services
	noteRepo := mongo.NewNoteRepo(mongoClient.Database("tododb"))
	notificationRepo := mongo.NewNotificationRepo(mongoClient.Database("tododb"))
	taskRepo := postgres.NewTaskRepo(db)

	taskCache := cache.New(rdb, 5*time.Minute)
	taskSvc := task.NewService(taskRepo, noteRepo, taskCache, producer, log)
	noteSvc := note.NewService(noteRepo, taskRepo)
	noteHandler := restapi.NewNoteHandler(noteSvc)
	notificationSvc := notification.NewService(notificationRepo, log)
	notificationHandler := restapi.NewNotificationHandler(notificationSvc)
	healthHandler := restapi.NewHealthHandler(db, mongoClient, rdb, rabbitConn, log)
	taskHandler := restapi.NewTaskHandler(taskSvc, log)

	// Router
	mux := http.NewServeMux()

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
	mux.Handle("GET /api/notifications", protected(notificationHandler.List))

	// HTTP Server
	srv := &http.Server{
		Addr:         cfg.TasksServerPort,
		Handler:      middleware.Logger(log, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("http server started", slog.String("addr", cfg.TasksServerPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-signalCtx.Done()

	log.Info("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	log.Info("server stopped")
}
