package exercise

type CreateExerciseDTO struct {
	Name        string        `json:"name"`
	Sets        []ExerciseSet `json:"sets"`
	Description string        `json:"description,omitempty"` // Описание упражнения, если нужно
}
