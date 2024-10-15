package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fit-journal/internal/apperror"
	"fit-journal/internal/auth"
	"fit-journal/internal/handlers"
	"fit-journal/pkg/logging"
	repeatable "fit-journal/pkg/utils"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

const (
	usersURL    = "/users"
	userURL     = "/users"
	registerURL = "/auth/register"
	loginURL    = "/auth/login"
)

type handler struct {
	logger     *logging.Logger
	repository Repository
}

func NewHandler(logger *logging.Logger, repo Repository) handlers.Handler {
	return &handler{
		logger:     logger,
		repository: repo,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	// Маршруты, не требующие аутентификации
	router.HandlerFunc(http.MethodPost, registerURL, apperror.Middleware(h.RegisterUser))
	router.HandlerFunc(http.MethodPost, loginURL, apperror.Middleware(h.Login))

	// Защищенные маршруты
	router.HandlerFunc(http.MethodGet, userURL, apperror.Middleware(auth.TokenAuthMiddleware(h.GetUserByUsername))) // Получить пользователя по UUID
	router.HandlerFunc(http.MethodPut, userURL, apperror.Middleware(auth.TokenAuthMiddleware(h.UpdateUser)))        // Обновить пользователя
	router.HandlerFunc(http.MethodDelete, userURL, apperror.Middleware(auth.TokenAuthMiddleware(h.DeleteUser)))     // Удалить пользователя
}

func (h *handler) RegisterUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Registering new user")

	var reqBody CreateUserDTO

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Invalid registration data", "", http.StatusBadRequest)
	}

	// Проверка обязательных полей
	err := repeatable.ValidateRequiredFields(map[string]string{
		"username": reqBody.Username,
		"password": reqBody.Password,
	})
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, err.Error(), "", http.StatusBadRequest)
	}

	// Проверяем, существует ли пользователь с таким же именем
	ctx := context.Background()
	foundUser, errFindOne := h.repository.FindOne(ctx, reqBody.Username)
	if foundUser.Username != "" {
		// Если пользователь найден, возвращаем ошибку
		h.logger.Error("User with this username already exists")
		return apperror.NewAppError(nil, "User with this username already exists", "", http.StatusConflict) // Conflict 409
	} else if errFindOne != nil && errFindOne.Error() != "no rows in result set" {
		h.logger.Info(errFindOne)
		return apperror.NewAppError(err, "Failed", "", http.StatusInternalServerError)
	}

	// Если пользователь не найден и нет ошибки, продолжаем регистрацию

	// Хэшируем пароль
	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Password hashing error", "", http.StatusInternalServerError)
	}

	newUser := User{
		Username:     reqBody.Username,
		PasswordHash: hashedPassword,
		BirthDate:    reqBody.BirthDate,
		Height:       reqBody.Height,
	}

	// Создаем пользователя в базе данных
	if err := h.repository.Create(ctx, newUser); err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Failed to save user", "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(newUser)
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("User login")

	var reqBody struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Invalid login data", "", http.StatusBadRequest)
	}

	// Проверка обязательных полей
	err := repeatable.ValidateRequiredFields(map[string]string{
		"username": reqBody.Username,
		"password": reqBody.Password,
	})
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, err.Error(), "", http.StatusBadRequest)
	}

	// Получаем пользователя из базы данных по имени пользователя
	ctx := context.Background()
	user, err := h.repository.FindOne(ctx, reqBody.Username)
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "User not found", "", http.StatusNotFound)
	}
	// Проверяем пароль
	if !auth.CheckPasswordHash(reqBody.Password, user.PasswordHash) {
		return apperror.NewAppError(nil, "Invalid login or password!", "", http.StatusUnauthorized)
	}

	// Генерируем JWT токен
	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Failed to generate token", "", http.StatusInternalServerError)
	}

	// Возвращаем токен
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func (h *handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Fetching user by username")

	// Извлекаем username из контекста, например, расшифрованный из JWT
	username, ok := r.Context().Value("username").(string)
	if !ok || strings.TrimSpace(username) == "" {
		return apperror.NewAppError(nil, "Invalid or missing username", "", http.StatusBadRequest)
	}

	ctx := context.Background()
	usr, err := h.repository.FindOne(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.logger.Error("User not found")
			return apperror.NewAppError(nil, "User not found", "", http.StatusNotFound)
		}
		h.logger.Error(err)
		return apperror.NewAppError(err, "Failed to fetch user", "", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(usr)
}

func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Updating user")

	// Получаем nickname пользователя из контекста
	nickname, ok := r.Context().Value("username").(string)
	if !ok || strings.TrimSpace(nickname) == "" {
		return apperror.NewAppError(nil, "Invalid or missing nickname in context", "", http.StatusUnauthorized)
	}

	// Ищем пользователя по никнейму
	ctx := context.Background()
	existingUser, err := h.repository.FindOne(ctx, nickname)
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "User not found", "", http.StatusNotFound)
	}

	// Парсим входящие данные для обновления
	var updates User
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Invalid user data", "", http.StatusBadRequest)
	}

	// Обновляем только те поля, которые были переданы
	if updates.Username != "" && updates.Username != existingUser.Username {
		existingUser.Username = updates.Username
	}
	if updates.BirthDate != "" && updates.BirthDate != existingUser.BirthDate {
		existingUser.BirthDate = updates.BirthDate
	}
	if updates.Height != "" && updates.Height != existingUser.Height {
		existingUser.Height = updates.Height
	}
	if updates.PasswordHash != "" {
		// Хэшируем новый пароль, если он был передан
		hashedPassword, err := auth.HashPassword(updates.PasswordHash)
		if err != nil {
			h.logger.Error(err)
			return apperror.NewAppError(err, "Failed to hash password", "", http.StatusInternalServerError)
		}
		existingUser.PasswordHash = hashedPassword
	}

	// Выполняем обновление в базе данных
	if err := h.repository.Update(ctx, existingUser); err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Failed to update user", "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	h.logger.Info("Deleting user")
	// Получаем nickname пользователя из контекста
	nickname, ok := r.Context().Value("username").(string)
	if !ok || strings.TrimSpace(nickname) == "" {
		return apperror.NewAppError(nil, "Invalid or missing nickname in context", "", http.StatusUnauthorized)
	}

	// Ищем пользователя по никнейму
	ctx := context.Background()
	existingUser, err := h.repository.FindOne(ctx, nickname)
	if err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "User not found", "", http.StatusNotFound)
	}

	if err := h.repository.Delete(ctx, existingUser.Username); err != nil {
		h.logger.Error(err)
		return apperror.NewAppError(err, "Failed to delete user", "", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
