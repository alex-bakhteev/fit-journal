package user

import (
	"context"
	"fit-journal/internal/entities/user"
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

// Create создает нового пользователя в БД
func (r *Repository) Create(ctx context.Context, user user.User) error {
	q := `
        INSERT INTO users
            (username, password_hash, birth_date, height)
        VALUES
            ($1, $2, $3, $4)
        RETURNING id
    `
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	// Сканируем ID в user.ID
	if err := r.client.QueryRow(ctx, q, user.Username, user.PasswordHash, user.BirthDate, user.Height).Scan(&user.ID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return newErr
		}
		return err
	}

	return nil
}

// FindAll возвращает список всех пользователей
func (r *Repository) FindAll(ctx context.Context) ([]user.User, error) {
	q := `
		SELECT id, username, birth_date, height FROM users;
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	rows, err := r.client.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]user.User, 0)

	for rows.Next() {
		var u user.User

		err = rows.Scan(&u.ID, &u.Username, &u.BirthDate, &u.Height)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FindOne ищет пользователя по ID
func (r *Repository) FindOne(ctx context.Context, username string) (user.User, error) {
	q := `
		SELECT id, username, password_hash, birth_date, height FROM users WHERE username = $1 AND is_deleted = FALSE
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	var u user.User
	err := r.client.QueryRow(ctx, q, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.BirthDate, &u.Height)
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}

// Update обновляет информацию о пользователе
func (r *Repository) Update(ctx context.Context, user user.User) error {
	q := `
		UPDATE users
		SET username = $1, password_hash = $2, birth_date = $3, height = $4
		WHERE id = $5  AND is_deleted = FALSE
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	_, err := r.client.Exec(ctx, q, user.Username, user.PasswordHash, user.BirthDate, user.Height, user.ID)
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

// Delete удаляет пользователя по ID
func (r *Repository) Delete(ctx context.Context, username string) error {
	q := `
		UPDATE users SET is_deleted = TRUE WHERE username = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(q)))

	_, err := r.client.Exec(ctx, q, username)
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
func NewRepository(client postgresql.Client, logger *logging.Logger) user.Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}
