package db

import (
	"context"
	"fit-journal/internal/entities/workout"
	"fit-journal/pkg/client/postgresql"
	"fit-journal/pkg/logging"
	"fmt"
	"github.com/jackc/pgconn"
	"strings"
)

type Repository struct {
	client postgresql.Client
	logger *logging.Logger
}

// formatQuery убирает переносы строк и табуляции из SQL-запроса для удобства логирования
func formatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

// Create создает новую тренировку в БД
func (r *Repository) Create(ctx context.Context, workout workout.Workout) (int64, error) {
	var id int64
	q := `
        INSERT INTO workouts
            (user_id, start_time, exercises)
        VALUES
            ($1, $2, $3)
        RETURNING id
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	// Сканируем ID в переменную id
	if err := r.client.QueryRow(ctx, q, workout.UserID, workout.StartTime, workout.Exercises).Scan(&id); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return 0, newErr
		}
		return 0, err
	}

	return id, nil
}

// FindAllByUserID возвращает список всех тренировок для конкретного пользователя
func (r *Repository) FindAllByUserID(ctx context.Context, userID int64) ([]workout.Workout, error) {
	q := `
		SELECT id, user_id, start_time, exercises FROM workouts WHERE user_id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	rows, err := r.client.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}

	workouts := make([]workout.Workout, 0)

	for rows.Next() {
		var w workout.Workout
		if err := rows.Scan(&w.ID, &w.UserID, &w.StartTime, &w.Exercises); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workouts, nil
}

// FindOne ищет тренировку по ID
func (r *Repository) FindOne(ctx context.Context, id int64) (workout.Workout, error) {
	q := `
		SELECT id, user_id, start_time, exercises FROM workouts WHERE id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	var w workout.Workout
	err := r.client.QueryRow(ctx, q, id).Scan(&w.ID, &w.UserID, &w.StartTime, &w.Exercises)
	if err != nil {
		return workout.Workout{}, err
	}

	return w, nil
}

// Update обновляет информацию о тренировке
func (r *Repository) Update(ctx context.Context, workout workout.Workout) error {
	q := `
		UPDATE workouts
		SET user_id = $1, start_time = $2, exercises = $3
		WHERE id = $4
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	_, err := r.client.Exec(ctx, q, workout.UserID, workout.StartTime, workout.Exercises, workout.ID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return newErr
		}
		return err
	}

	return nil
}

// Delete удаляет тренировку по ID
func (r *Repository) Delete(ctx context.Context, id int64) error {
	q := `
		DELETE FROM workouts
		WHERE id = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	_, err := r.client.Exec(ctx, q, id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return newErr
		}
		return err
	}

	return nil
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(client postgresql.Client, logger *logging.Logger) *Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}
