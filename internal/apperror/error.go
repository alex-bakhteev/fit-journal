package apperror

import "encoding/json"

type AppError struct {
	Err              error  `json:"-"`
	Message          string `json:"message,omitempty"`
	DeveloperMessage string `json:"developer_message,omitempty"`
	StatusCode       int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) Marshal() []byte {
	marshal, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return marshal
}

// NewAppError создает ошибку с автоматическим использованием стандартного статус-кода
func NewAppError(err error, message, developerMessage string, statusCode int) *AppError {
	// Если статус-код не передан (или 0), используем стандартный 500 Internal Server Error
	if statusCode == 0 {
		statusCode = 500
	}
	return &AppError{
		Err:              err,
		Message:          message,
		DeveloperMessage: developerMessage,
		StatusCode:       statusCode,
	}
}

// systemError создает ошибку системы, используя стандартные механизмы
func systemError(err error) *AppError {
	return NewAppError(err, "internal system error", err.Error(), 500)
}
