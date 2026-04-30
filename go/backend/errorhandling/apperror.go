package errorhandling

import (
	"fmt"
)

type AppError struct {
	ErrorCode     string
	Message       string
	Fields        []AppErrorField
	Preconditions []AppErrorPrecondition
}

type AppErrorField struct {
	Field string
	Error string
}

type AppErrorPrecondition struct {
	ErrorCode   string
	Subject     string
	Description string
}

// TODO check
func (e *AppError) Error() string {
	return fmt.Sprintf("AppError: code=%s msg=%s fields=%v", e.ErrorCode, e.Message, e.Fields)
}

func NewInternalError(msg string) *AppError {
	return &AppError{
		ErrorCode: ErrCodeInternal,
		Message:   msg,
	}
}

func NewBadRequestError(msg string) *AppError {
	return &AppError{
		ErrorCode: ErrCodeBadRequest,
		Message:   msg,
	}
}

// func newAppError(errCode AppErrorCode, msg string, fields []AppErrorField) error {

// 	appError := &AppError{
// 		ErrorCode: errCode,
// 		Message:   msg,
// 		Fields:    fields,
// 	}

// 	return appError
// }

// func appendAppErrField(fields []AppErrorField, field string, description string) []AppErrorField {
// 	result := append(fields, newAppErrorField(field, description))
// 	return result
// }

// func newAppErrorField(field string, description string) AppErrorField {
// 	return AppErrorField{
// 		Field: field,
// 		Error: description,
// 	}
// }
