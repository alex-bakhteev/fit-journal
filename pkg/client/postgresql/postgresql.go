package postgresql

import (
	"context"
	"fit-journal/internal/config"
	repeatable "fit-journal/pkg/utils"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"time"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewClient(ctx context.Context, maxAttempts int, sc config.StorageConfig) (pool *pgxpool.Pool, err error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", sc.Username, sc.Password, sc.Host, sc.Port, sc.Database)
	err = repeatable.DoWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = pgxpool.Connect(ctx, dsn)
		if err != nil {
			return err
		}

		return nil
	}, maxAttempts, 5*time.Second)

	if err != nil {
		log.Fatal("error do with tries postgresql")
	}

	// Создание таблиц
	if err := createTables(ctx, pool); err != nil {
		return nil, err
	}

	return pool, nil
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	// Таблица пользователей
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		birth_date TEXT,
		height TEXT
	);`

	// Таблица упражнений
	exerciseTable := `
	CREATE TABLE IF NOT EXISTS exercises (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		sets JSONB NOT NULL,  -- Массив подходов сохраняется как JSON
		description TEXT
	);`

	// Таблица метрик пользователя
	metricTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		weight TEXT,
		calories_consumed TEXT,
		day TEXT NOT NULL
	);`

	// Таблица тренировок
	workoutTable := `
	CREATE TABLE IF NOT EXISTS workouts (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		start_time BIGINT NOT NULL,
		exercises JSONB NOT NULL  -- Список упражнений сохраняется как JSON
	);`

	// Выполнение SQL-запросов на создание таблиц
	tables := []string{userTable, exerciseTable, metricTable, workoutTable}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}
