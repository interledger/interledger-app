package external

import (
	"errors"
	"fmt"
)

var (
	ErrInternal = errors.New("unit client: internal error")
	ErrRequest  = errors.New("unit client: request error")
)

type ErrHttp struct {
	Code   int
	Errors []ResponseError
}

func (e *ErrHttp) Error() string {
	return fmt.Sprintf("unit client: http error. statusCode=%d, errors=%+v", e.Code, e.Errors)
}
