package pkg

import (
	"github.com/gofiber/fiber/v2"
)

// Response envelope structs used by all handlers.

type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

type ErrorBody struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.JSON(SuccessResponse{Success: true, Data: data})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Success: true, Data: data})
}

func Message(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(MessageResponse{Success: true, Message: message})
}

func BadRequestResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "BAD_REQUEST", Message: message},
	})
}

func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "UNAUTHORIZED", Message: message},
	})
}

func ForbiddenResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "FORBIDDEN", Message: message},
	})
}

func NotFoundResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "NOT_FOUND", Message: message},
	})
}

func ConflictResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "CONFLICT", Message: message},
	})
}

func ValidationResponse(c *fiber.Ctx, message string, fields map[string][]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Fields:  fields,
		},
	})
}

func RateLimitedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "RATE_LIMITED", Message: message},
	})
}

func ServiceUnavailableResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
		Success: false,
		Error:   ErrorBody{Code: "SERVICE_UNAVAILABLE", Message: message},
	})
}

// Error is the central error responder. It inspects the error type and
// maps it to the proper HTTP response envelope.
func Error(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if ok := AsAppError(err, &appErr); ok {
		return writeAppError(c, appErr)
	}
	return writeAppError(c, Internal("internal server error", err))
}

func writeAppError(c *fiber.Ctx, appErr *AppError) error {
	body := ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Fields:  appErr.Details,
		},
	}
	return c.Status(appErr.HTTPStatus).JSON(body)
}

// AsAppError unwraps the chain to find an *AppError.
func AsAppError(err error, target **AppError) bool {
	for err != nil {
		if ae, ok := err.(*AppError); ok {
			*target = ae
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
