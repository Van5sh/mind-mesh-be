package apperrors

import "errors"

type Code string

const (
	InvalidArgument Code = "INVALID_ARGUMENT"
	NotFound        Code = "NOT_FOUND"
	Conflict        Code = "CONFLICT"
	Forbidden       Code = "FORBIDDEN"
	Unauthorized    Code = "UNAUTHORIZED"
	Internal        Code = "INTERNAL"
)

type AppError struct {
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func New(code Code, message string, err error) error {
	return &AppError{Code: code, Message: message, Err: err}
}

func Validation(message string) error       { return New(InvalidArgument, message, nil) }
func NotFoundError(message string) error    { return New(NotFound, message, nil) }
func ConflictError(message string) error    { return New(Conflict, message, nil) }
func ForbiddenError(message string) error   { return New(Forbidden, message, nil) }
func UnauthorizedError(message string) error { return New(Unauthorized, message, nil) }
func InternalError(message string, err error) error {
	return New(Internal, message, err)
}

func CodeOf(err error) Code {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return Internal
}

func IsCode(err error, code Code) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == code
}
