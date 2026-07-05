// Package utils — helpers compartidos.
// validation.go: Gin valida contra los tags del DTO al hacer bind; este helper
// convierte los fallos en la respuesta estándar ({field, message}).
package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"idilica-backend-go/internal/apperrors"
)

// init registers the json tag as the reported field name, so validation
// details say "username" (the API field) instead of "Username" (the Go field).
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			for _, tag := range []string{"json", "uri"} {
				name := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]
				if name != "" && name != "-" {
					return name
				}
			}
			return field.Name
		})
	}
}

// BindJSON bindea y valida el body: utils.BindJSON[dto.LoginDto](c).
func BindJSON[T any](c *gin.Context) (*T, error) {
	var obj T
	if err := c.ShouldBindJSON(&obj); err != nil {
		return nil, toValidationError(err)
	}
	return &obj, nil
}

// BindUri bindea y valida los parámetros de ruta.
func BindUri[T any](c *gin.Context) (*T, error) {
	var obj T
	if err := c.ShouldBindUri(&obj); err != nil {
		return nil, toValidationError(err)
	}
	return &obj, nil
}

func toValidationError(err error) error {
	var vErrs validator.ValidationErrors
	if errors.As(err, &vErrs) {
		details := make([]map[string]string, 0, len(vErrs))
		for _, fe := range vErrs {
			details = append(details, map[string]string{
				"field":   fe.Field(),
				"message": messageFor(fe),
			})
		}
		return apperrors.NewValidation("Validation failed", details)
	}
	// Malformed JSON, wrong types, etc.
	return apperrors.NewBadRequest("Invalid request payload")
}

// messageFor produce mensajes legibles por tag de validación.
func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%q is required", fe.Field())
	case "email":
		return fmt.Sprintf("%q must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("%q must be at least %s characters or greater than %s", fe.Field(), fe.Param(), fe.Param())
	case "max":
		return fmt.Sprintf("%q must be at most %s characters or less than %s", fe.Field(), fe.Param(), fe.Param())
	case "len":
		return fmt.Sprintf("%q must be exactly %s characters", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%q must be a valid UUID", fe.Field())
	default:
		return fmt.Sprintf("%q is invalid (%s)", fe.Field(), fe.Tag())
	}
}
