package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"task1/internal/cache"
	"task1/internal/config"
	"task1/internal/order"
	"task1/internal/order/db"
	"task1/internal/serv"
	"task1/pkg/client"
	"task1/pkg/migr"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

var errChan = make(chan error, 2)

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGINT)
	logger := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		logger.Error("Ошибка загрузки переменных окружение .env", "error", err.Error())
		err = godotenv.Load("example.env")
		if err != nil {
			logger.Error("Ошибка загрузки переменных окружение example.env", "error", err.Error())
			errChan <- err
		}

	}
	cfg, err := config.MustLoad()
	if err != nil {
		logger.Error("Ошибка при загрузке конфига", "error", err)
		errChan <- err
	}
	dbCLient, err := client.NewCLient(ctx, logger)
	if err != nil {
		logger.Error("Ошибка подключение к бд", "error", err.Error())
		errChan <- err
	}
	migrator := migr.Migrator{
		Pool:   dbCLient,
		Logger: logger,
	}
	err = migrator.Migrate(os.Getenv("MIGRATE_PATH"))
	if err != nil {
		logger.Error("Ошибка миграции бд", "error", err)
		errChan <- err
	}
	reader := createKafkaReader()
	defer reader.Close()
	repository := db.NewRepository(dbCLient, logger)
	cacheForOrders := cache.NewOrderCache(logger)
	cacheForOrders.RestoreFromDB(ctx, repository)
	server := serv.NewServer(*cacheForOrders, logger, repository, cfg.Port)
	go server.Start()
	go readMessageFromKafka(ctx, reader, logger, repository)
	gracefulShutdown := func() {
		logger.Info("GRACEFUL SHUTDOWN")
		close(errChan)
		close(stopChan)
		dbCLient.Close()
		server.Stop(ctx)
		cancel()
	}
	select {
	case <-errChan:
		gracefulShutdown()
		os.Exit(0)

	case <-stopChan:
		gracefulShutdown()
		os.Exit(0)

	}
}

func createKafkaReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "my-topic",
		GroupID: "group-1",
	})
}

func readMessageFromKafka(ctx context.Context, reader *kafka.Reader, logger *slog.Logger, repository order.Repository) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error("Ошибка при получении", "error", err)
				errChan <- err
			}
			var ord order.Order
			if err := json.Unmarshal(msg.Value, &ord); err != nil {
				logger.Error("Ошибка парсинга сообщения", "error", err)
				continue
			}
			fmt.Println(ord)
			ord.OrderUID, err = repository.Save(ctx, ord)
			if err != nil {
				logger.Error("Ошибка при сохранении в бд", "error", err)
			}
			logger.Info("Получен заказ")
		}
	}
}
