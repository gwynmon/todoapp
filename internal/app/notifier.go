package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"todoapp/internal/entity"

	"todoapp/config"
	mongorepo "todoapp/internal/repository/mongo"
	notificationUC "todoapp/internal/usecases/notification"
	"todoapp/pkg/broker"
	"todoapp/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	mongoDriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func RunNotifier(cfg *config.Config) {
	log := logger.Init(cfg.LogLevel)

	ctx := context.Background()

	mongoClient, err := mongoDriver.Connect(
		options.Client().ApplyURI(cfg.MongoDSN),
	)
	if err != nil {
		log.Error("mongo connection failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())

	log.Info("mongodb connected successfully")

	notificationRepo := mongorepo.NewNotificationRepo(
		mongoClient.Database("tododb"),
	)

	notificationSvc := notificationUC.NewService(
		notificationRepo,
		log,
	)

	rabbitConn, err := amqp.Dial(cfg.RabbitMQDSN)
	if err != nil {
		log.Error("rabbitmq connection failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	log.Info("rabbitmq connected successfully")

	consumer, err := broker.NewConsumer(rabbitConn)
	if err != nil {
		log.Error(
			"rabbitmq consumer failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer consumer.Close()

	messages, err := consumer.Consume()
	if err != nil {
		log.Error(
			"consume failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	go func() {
		for msg := range messages {

			var event entity.TaskEvent

			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Error(
					"event unmarshal failed",
					slog.String("error", err.Error()),
				)
				continue
			}

			if err := notificationSvc.CreateFromEvent(
				context.Background(),
				event,
			); err != nil {

				log.Error(
					"notification save failed",
					slog.String("error", err.Error()),
				)
				continue
			}

			log.Info(
				"notification saved",
				slog.String("event_type", event.EventType),
				slog.Int64("task_id", event.TaskID),
			)
		}
	}()

	log.Info("notifier service started")

	signalCtx, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-signalCtx.Done()

	log.Info("notifier service stopped")
}
