package validators

import (
	"github.com/go-playground/validator"
)

type ValidationResponse struct {
	Status string            `json:"status"`
	Errors map[string]string `json:"errors"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func Error(msg string) ValidationResponse {
	return ValidationResponse{
		Status: StatusError,
		Errors: map[string]string{"error": msg},
	}
}

func OK() ValidationResponse {
	return ValidationResponse{
		Status: StatusOK,
	}
}

func ValidationError(errs validator.ValidationErrors) ValidationResponse {
	errorsMap := make(map[string]string)

	for _, err := range errs {
		fieldName := err.Field()
		tag := err.Tag()

		var message string

		switch tag {
		case "required":
			message = "Это поле обязательно"
		}

		errorsMap[fieldName] = message
	}

	return ValidationResponse{
		Status: StatusError,
		Errors: errorsMap,
	}
}
