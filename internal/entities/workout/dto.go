package workout

import (
	"fit-journal/internal/entities/exercise"
	"fit-journal/internal/entities/user"
)

type CreateWorkoutDTO struct {
	User      user.User           `json:"user"`       // Пользователь, который выполняет тренировку
	StartTime int64               `json:"start_time"` // Время начала тренировки (Unix timestamp)
	Exercises []exercise.Exercise `json:"exercises"`  // Список упражнений в тренировке
}
