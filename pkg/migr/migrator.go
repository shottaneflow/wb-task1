package migr

import (
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Migrator struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

func NewMigrator(pool *pgxpool.Pool, logger *slog.Logger) *Migrator {
	return &Migrator{Pool: pool, Logger: logger}
}

func (m *Migrator) Migrate(path string) error {
	db := stdlib.OpenDBFromPool(m.Pool)
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		m.Logger.Error("Ошибка при создании драйвера бд", "error", err)
		return err
	}
	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
		"postgres", driver)
	if err != nil {
		m.Logger.Error("Ошибка при создании миграции", "error", err)
		return err
	}
	err = migrator.Up()
	defer migrator.Close()
	if err != nil && err != migrate.ErrNoChange {
		m.Logger.Error("Ошибка миграции", "error", err)
		return err
	}
	m.Logger.Info("Миграция выполнена")

	return nil
}
