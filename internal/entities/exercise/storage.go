package exercise

import "context"

type Repository interface {
	Create(ctx context.Context, exercise Exercise) error
	FindOne(ctx context.Context, id string) (Exercise, error)
	Update(ctx context.Context, exercise Exercise) error
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context) (e []Exercise, err error)
}
