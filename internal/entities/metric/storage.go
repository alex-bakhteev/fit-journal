package metric

import "context"

type Repository interface {
	Create(ctx context.Context, metric Metric) error
	FindOne(ctx context.Context, id string) (Metric, error)
	Update(ctx context.Context, metric Metric) error
	Delete(ctx context.Context, id string) error
	FindAll(ctx context.Context) (m []Metric, err error)
}
