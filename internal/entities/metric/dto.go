package metric

import (
	"fit-journal/internal/entities/user"
)

type CreateMetricDTO struct {
	UserID           user.User `json:"id"`
	Weigth           string    `json:"weigth,omitempty"`
	CaloriesConsumed string    `json:"calories_consumed,omitempty"`
	Day              string    `json:"day"`
}
