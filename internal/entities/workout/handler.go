package workout

import (
	"encoding/json"
	"fit-journal/internal/apperror"
	"fit-journal/internal/auth"
	"fit-journal/internal/entities/exercise"
	"fit-journal/internal/entities/user"
	"fit-journal/internal/handlers"
	"fit-journal/pkg/logging"
	"github.com/julienschmidt/httprouter"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	workoutsURL = "/workouts"
	workoutURL  = "/workouts/:workout_id"
	exerciseURL = "/workouts/:workout_id/exercises/:exercise_id"
	setURL      = "/workouts/:workout_id/exercises/:exercise_id/sets/:set_id"
)

type handler struct {
	logger         *logging.Logger
	repository     Repository
	userRepository user.Repository
}

func NewHandler(logger *logging.Logger, repo Repository, userRepo user.Repository) handlers.Handler {
	return &handler{
		logger:         logger,
		repository:     repo,
		userRepository: userRepo,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, workoutsURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.CreateWorkout))))
	router.HandlerFunc(http.MethodPut, workoutURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.UpdateWorkout))))
	router.HandlerFunc(http.MethodGet, workoutURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.GetWorkoutByID))))
	router.HandlerFunc(http.MethodGet, workoutsURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.GetAllWorkouts))))
	router.HandlerFunc(http.MethodPost, exerciseURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.AddSetToExercise))))
	router.HandlerFunc(http.MethodDelete, workoutURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.DeleteWorkout))))
	router.HandlerFunc(http.MethodDelete, exerciseURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.DeleteExercise))))
	router.HandlerFunc(http.MethodDelete, setURL, apperror.Middleware(auth.TokenAuthMiddleware(apperror.AppHandler(h.DeleteSet))))
}

func (h *handler) CreateWorkout(w http.ResponseWriter, r *http.Request) error {
	// Извлечение username из контекста
	username, ok := r.Context().Value("username").(string)
	if !ok {
		h.logger.Error("Ошибка извлечения username из контекста")
		return apperror.NewAppError(nil, "Ошибка аутентификации", "Не удалось получить пользователя", http.StatusUnauthorized)
	}

	// Получение user_id на основе username
	user, err := h.userRepository.FindOne(r.Context(), username)
	if err != nil {
		h.logger.Error("Ошибка получения пользователя по username: %v", err)
		return apperror.NewAppError(err, "Ошибка при создании тренировки", "Ошибка получения пользователя", http.StatusInternalServerError)
	}
	// Создание новой тренировки с пустым списком упражнений и текущей датой
	workout := Workout{
		UserID:    user.ID,
		StartTime: time.Now().Unix(),
		Exercises: []exercise.Exercise{},
	}

	// Вызов репозитория для создания тренировки
	id, err := h.repository.Create(r.Context(), workout)
	if err != nil {
		h.logger.Error("Ошибка создания тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при создании тренировки", "Ошибка взаимодействия с базой данных", http.StatusInternalServerError)
	}

	// Устанавливаем ID в workout
	workout.ID = id

	// Ответ с созданной тренировкой
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(workout); err != nil {
		return apperror.NewAppError(err, "Ошибка при отправке ответа", "Ошибка кодирования JSON", http.StatusInternalServerError)
	}

	return nil
}

// UpdateWorkout обновляет существующую тренировку, добавляя новое упражнение
func (h *handler) UpdateWorkout(w http.ResponseWriter, r *http.Request) error {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// Преобразуем строковый ID в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования id: %v", err)
		return apperror.NewAppError(err, "Неверный формат id", "Ошибка преобразования ID", http.StatusBadRequest)
	}

	// Декодируем данные нового упражнения
	var newExercise exercise.Exercise
	if err := json.NewDecoder(r.Body).Decode(&newExercise); err != nil {
		h.logger.Error("Ошибка декодирования тела запроса: %v", err)
		return apperror.NewAppError(err, "Неверный формат данных", "Ошибка декодирования JSON", http.StatusBadRequest)
	}

	// Получаем текущую тренировку
	ctx := r.Context()
	workout, err := h.repository.FindOne(ctx, id)
	if err != nil {
		h.logger.Error("Ошибка получения тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при добавлении упражнения", "Тренировка не найдена", http.StatusInternalServerError)
	}

	// Генерируем уникальный int64 ID для нового упражнения
	newExercise.ID = rand.Int63()

	// Добавляем новое упражнение в массив упражнений
	workout.Exercises = append(workout.Exercises, newExercise)

	// Обновляем тренировку в базе данных
	if err := h.repository.Update(ctx, workout); err != nil {
		h.logger.Error("Ошибка обновления тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при добавлении упражнения", "Ошибка взаимодействия с базой данных", http.StatusInternalServerError)
	}

	// Возвращаем обновленную тренировку
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(workout); err != nil {
		return apperror.NewAppError(err, "Ошибка при отправке ответа", "Ошибка кодирования JSON", http.StatusInternalServerError)
	}

	return nil
}

// GetWorkoutByID получает тренировку по ID
func (h *handler) GetWorkoutByID(w http.ResponseWriter, r *http.Request) error {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// Преобразуем строковый ID в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования id: %v", err)
		return apperror.NewAppError(err, "Неверный формат id", "Ошибка преобразования ID", http.StatusBadRequest)
	}

	// Получаем текущую тренировку
	ctx := r.Context()
	workout, err := h.repository.FindOne(ctx, id)
	if err != nil {
		h.logger.Error("Ошибка получения тренировки: %v", err)
		return apperror.NewAppError(err, "Тренировка не найдена", "Ошибка взаимодействия с базой данных", http.StatusNotFound)
	}

	// Возвращаем тренировку в ответ
	if err := json.NewEncoder(w).Encode(workout); err != nil {
		return apperror.NewAppError(err, "Ошибка при отправке ответа", "Ошибка кодирования JSON", http.StatusInternalServerError)
	}

	return nil
}

// GetAllWorkouts получает все тренировки для конкретного пользователя
func (h *handler) GetAllWorkouts(w http.ResponseWriter, r *http.Request) error {
	// Получаем username из контекста
	username, ok := r.Context().Value("username").(string)
	if !ok {
		h.logger.Error("Ошибка извлечения username из контекста")
		return apperror.NewAppError(nil, "Не удалось получить данные пользователя", "Ошибка контекста", http.StatusInternalServerError)
	}

	// Ищем user_id по username
	user, err := h.userRepository.FindOne(r.Context(), username)
	if err != nil {
		h.logger.Error("Ошибка получения пользователя по username: %v", err)
		return apperror.NewAppError(err, "Ошибка при получении тренировок", "Ошибка получения пользователя", http.StatusInternalServerError)
	}

	// Ищем все тренировки для найденного пользователя
	workouts, err := h.repository.FindAllByUserID(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("Ошибка получения тренировок: %v", err)
		return apperror.NewAppError(err, "Ошибка при получении тренировок", "Ошибка взаимодействия с базой данных", http.StatusInternalServerError)
	}

	// Отправляем ответ
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(workouts); err != nil {
		return apperror.NewAppError(err, "Ошибка при отправке ответа", "Ошибка кодирования JSON", http.StatusInternalServerError)
	}

	return nil
}

// AddSetToExercise добавляет новый подход к упражнению в тренировке
func (h *handler) AddSetToExercise(w http.ResponseWriter, r *http.Request) error {
	// Получаем ID тренировки и ID упражнения из параметров URL
	workoutIDStr := httprouter.ParamsFromContext(r.Context()).ByName("workout_id")
	exerciseIDStr := httprouter.ParamsFromContext(r.Context()).ByName("exercise_id")

	// Преобразуем строковые ID в int64
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования workout_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат workout_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}
	exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования exercise_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат exercise_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}

	// Декодируем новый подход (с весом и повторами)
	var newSet exercise.ExerciseSet
	if err := json.NewDecoder(r.Body).Decode(&newSet); err != nil {
		h.logger.Error("Ошибка декодирования тела запроса: %v", err)
		return apperror.NewAppError(err, "Неверный формат данных", "Ошибка декодирования JSON", http.StatusBadRequest)
	}

	// Присваиваем уникальный ID для нового подхода
	newSet.ID = rand.Int63()

	// Получаем текущую тренировку
	ctx := r.Context()
	workout, err := h.repository.FindOne(ctx, workoutID)
	if err != nil {
		h.logger.Error("Ошибка получения тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при добавлении подхода", "Тренировка не найдена", http.StatusInternalServerError)
	}

	// Ищем упражнение по его ID в тренировке
	foundExercise := false
	for i := range workout.Exercises {
		if workout.Exercises[i].ID == exerciseID {
			// Добавляем новый подход к упражнению
			workout.Exercises[i].Sets = append(workout.Exercises[i].Sets, newSet)
			foundExercise = true
			break
		}
	}

	// Если упражнение не найдено, возвращаем ошибку
	if !foundExercise {
		h.logger.Error("Упражнение не найдено в тренировке")
		return apperror.NewAppError(nil, "Упражнение не найдено", "Ошибка поиска упражнения в тренировке", http.StatusNotFound)
	}

	// Обновляем тренировку в базе данных
	if err := h.repository.Update(ctx, workout); err != nil {
		h.logger.Error("Ошибка обновления тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при добавлении подхода", "Ошибка взаимодействия с базой данных", http.StatusInternalServerError)
	}

	// Возвращаем обновленную тренировку
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(workout); err != nil {
		return apperror.NewAppError(err, "Ошибка при отправке ответа", "Ошибка кодирования JSON", http.StatusInternalServerError)
	}

	return nil
}

// DeleteWorkout удаляет тренировку по её ID
func (h *handler) DeleteWorkout(w http.ResponseWriter, r *http.Request) error {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("workout_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования id: %v", err)
		return apperror.NewAppError(err, "Неверный формат id", "Ошибка преобразования ID", http.StatusBadRequest)
	}

	// Удаление тренировки
	if err := h.repository.Delete(r.Context(), id); err != nil {
		h.logger.Error("Ошибка удаления тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при удалении тренировки", "Тренировка не найдена", http.StatusNotFound)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

// DeleteExercise удаляет упражнение из тренировки по ID
func (h *handler) DeleteExercise(w http.ResponseWriter, r *http.Request) error {
	workoutIDStr := httprouter.ParamsFromContext(r.Context()).ByName("workout_id")
	exerciseIDStr := httprouter.ParamsFromContext(r.Context()).ByName("exercise_id")
	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования workout_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат workout_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}
	exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования exercise_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат exercise_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}

	// Получаем текущую тренировку
	ctx := r.Context()
	workout, err := h.repository.FindOne(ctx, workoutID)
	if err != nil {
		h.logger.Error("Ошибка получения тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при удалении упражнения", "Тренировка не найдена", http.StatusInternalServerError)
	}

	// Удаляем упражнение
	var updatedExercises []exercise.Exercise
	found := false
	for _, ex := range workout.Exercises {
		if ex.ID != exerciseID {
			updatedExercises = append(updatedExercises, ex)
		} else {
			found = true
		}
	}
	if !found {
		return apperror.NewAppError(nil, "Упражнение не найдено", "Ошибка поиска упражнения", http.StatusNotFound)
	}
	workout.Exercises = updatedExercises

	// Обновляем тренировку в базе данных
	if err := h.repository.Update(ctx, workout); err != nil {
		h.logger.Error("Ошибка обновления тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при удалении упражнения", "Ошибка взаимодействия с базой данных", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

// DeleteSet удаляет подход из упражнения
func (h *handler) DeleteSet(w http.ResponseWriter, r *http.Request) error {
	workoutIDStr := httprouter.ParamsFromContext(r.Context()).ByName("workout_id")
	exerciseIDStr := httprouter.ParamsFromContext(r.Context()).ByName("exercise_id")
	setIDStr := httprouter.ParamsFromContext(r.Context()).ByName("set_id")

	workoutID, err := strconv.ParseInt(workoutIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования workout_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат workout_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}
	exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования exercise_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат exercise_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}
	setID, err := strconv.ParseInt(setIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Ошибка преобразования set_id: %v", err)
		return apperror.NewAppError(err, "Неверный формат set_id", "Ошибка преобразования ID", http.StatusBadRequest)
	}

	// Получаем текущую тренировку
	ctx := r.Context()
	workout, err := h.repository.FindOne(ctx, workoutID)
	if err != nil {
		h.logger.Error("Ошибка получения тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при удалении подхода", "Тренировка не найдена", http.StatusInternalServerError)
	}

	// Находим упражнение
	foundExercise := false
	for i := range workout.Exercises {
		if workout.Exercises[i].ID == exerciseID {
			// Удаляем подход
			var updatedSets []exercise.ExerciseSet
			foundSet := false
			for _, set := range workout.Exercises[i].Sets {
				if set.ID != setID {
					updatedSets = append(updatedSets, set)
				} else {
					foundSet = true
				}
			}
			if !foundSet {
				return apperror.NewAppError(nil, "Подход не найден", "Ошибка поиска подхода", http.StatusNotFound)
			}
			workout.Exercises[i].Sets = updatedSets
			foundExercise = true
			break
		}
	}

	if !foundExercise {
		return apperror.NewAppError(nil, "Упражнение не найдено", "Ошибка поиска упражнения", http.StatusNotFound)
	}

	// Обновляем тренировку в базе данных
	if err := h.repository.Update(ctx, workout); err != nil {
		h.logger.Error("Ошибка обновления тренировки: %v", err)
		return apperror.NewAppError(err, "Ошибка при удалении подхода", "Ошибка взаимодействия с базой данных", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
