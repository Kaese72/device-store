package systemerrors

import "net/http"

type SystemError interface {
	Error() string
	Reason() int
}

type Reason int

const (
	NotFound      Reason = http.StatusNotFound
	UserError     Reason = http.StatusBadRequest
	InternalError Reason = http.StatusInternalServerError
)

type apiError struct {
	err  error
	code Reason
}

func (err apiError) Reason() int {
	return int(err.code)
}

func (err apiError) Error() string {
	return err.err.Error()
}

func WrapSystemError(err error, intention Reason) apiError {
	return apiError{
		err:  err,
		code: intention,
	}
}
