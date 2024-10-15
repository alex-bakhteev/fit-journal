package apperror

import (
	"errors"
	"net/http"
)

// Определяем ошибку ErrNotFound
var ErrNotFound = NewAppError(errors.New("not found"), "Resource not found", "The requested resource could not be found", http.StatusNotFound)

// AppHandler — это тип обработчиков, которые возвращают ошибку
type AppHandler func(w http.ResponseWriter, r *http.Request) error

// Middleware оборачивает обработчик с возвратом ошибки в http.HandlerFunc
func Middleware(h AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var appErr *AppError
		err := h(w, r) // вызываем исходный обработчик

		if err != nil {
			w.Header().Set("Content-Type", "application/json")

			// Если ошибка является кастомной (AppError)
			if errors.As(err, &appErr) {
				// Обработка 404 Not Found
				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(ErrNotFound.Marshal())
					return
				}

				// Обработка остальных кастомных ошибок
				w.WriteHeader(appErr.StatusCode)
				w.Write(appErr.Marshal())
				return
			}

			// Для системных ошибок возвращаем 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(systemError(err).Marshal())
		}
	}
}
