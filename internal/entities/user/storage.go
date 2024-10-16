package user

import "context"

type Repository interface {
	Create(ctx context.Context, user User) error
	FindOne(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context) (u []User, err error)
}
