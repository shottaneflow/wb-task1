package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"task1/internal/client"
	"task1/pkg/migr"

	"github.com/joho/godotenv"
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
		logger.Error("Ошибка загрузки переменных окружение", "error", err.Error())
		errChan <- err
	}
	dbCLient, err := client.NewCLient(ctx, logger)
	if err != nil {
		logger.Error("Ошибка подключение к бд", "error", err.Error())
		errChan <- err
	}
	migrator:=migr.Migrator{
		Pool: dbCLient,
	}
	err=migrator.Migrate(os.Getenv("MIGRATE_PATH"))
	if err!=nil{
		logger.Error("Ошибка миграции бд","error",err)
		errChan<-err
	}

	gracefulShutdown := func() {
		logger.Info("GRACEFUL SHUTDOWN")
		close(errChan)
		close(stopChan)
		cancel()
		dbCLient.Close()
	}

	err = dbCLient.Ping(ctx)
	if err != nil {
		logger.Error("Ошибка при пинге бд", "error", err.Error())
		errChan <- err
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
