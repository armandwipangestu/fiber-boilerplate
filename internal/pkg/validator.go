package pkg

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validator validates structs against `validate` tags and returns field
// errors keyed by field name, suitable for the response envelope.
type Validator interface {
	ValidateStruct(s any) (map[string][]string, error)
}

var (
	validateOnce sync.Once
	validate     *validator.Validate
)

// NewValidator builds (once) the shared validator with custom tags registered.
func NewValidator() Validator {
	validateOnce.Do(func() {
		validate = validator.New()
		_ = validate.RegisterValidation("uuid", isUUIDValue)
	})
	return &structValidator{}
}

type structValidator struct{}

// ValidateStruct runs validation and normalizes errors to a field map.
func (v *structValidator) ValidateStruct(s any) (map[string][]string, error) {
	if err := validate.Struct(s); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			fields := make(map[string][]string, len(verrs))
			for _, fe := range verrs {
				field := strings.ToLower(fe.StructField())
				fields[field] = append(fields[field], validationMessage(fe))
			}
			return fields, fmt.Errorf("validation failed: %w", err)
		}
		return nil, err
	}
	return nil, nil
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "uuid":
		return "must be a valid UUID"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	default:
		return fmt.Sprintf("failed %s validation", fe.Tag())
	}
}

func isUUIDValue(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}
