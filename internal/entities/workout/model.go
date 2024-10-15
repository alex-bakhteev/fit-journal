package workout

import (
	"fit-journal/internal/entities/exercise"
)

type Workout struct {
	ID        int64               `json:"id"`
	UserID    int64               `json:"user_id"`
	StartTime int64               `json:"start_time"`
	Exercises []exercise.Exercise `json:"exercises"`
}
