package exercise

type Exercise struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Sets        []ExerciseSet `json:"sets"`
	Description string        `json:"description,omitempty"` // Описание упражнения, если нужно
}

type ExerciseSet struct {
	ID     int64   `json:"id"` // Уникальный ID для подхода
	Reps   int     `json:"reps"`
	Weight float64 `json:"weight"` // Вес для каждого подхода
}
