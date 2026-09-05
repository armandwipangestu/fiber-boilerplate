package user

import "github.com/armandwipangestu/fiber-boilerplate/internal/pkg"

// Feature-level business errors. These are *pkg.AppError instances so they
// carry a code, message, and HTTP status that the handler maps directly.
var (
	ErrUserNotFound     = pkg.NewAppError("USER_NOT_FOUND", "user not found", 404, nil)
	ErrEmailExists      = pkg.NewAppError("EMAIL_EXISTS", "email already exists", 409, nil)
	ErrCannotSelfDelete = pkg.NewAppError("BUSINESS_ERROR", "cannot delete your own account", 422, nil)
	ErrNoPermission     = pkg.NewAppError("FORBIDDEN", "insufficient permissions", 403, nil)
)
