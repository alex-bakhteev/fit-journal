package workout

import "context"

type Repository interface {
	Create(ctx context.Context, workout Workout) (int64, error)
	FindOne(ctx context.Context, id int64) (Workout, error)
	Update(ctx context.Context, workout Workout) error
	Delete(ctx context.Context, id int64) error
	FindAllByUserID(ctx context.Context, id int64) (w []Workout, err error)
}
