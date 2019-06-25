package api

import "github.com/pkg/errors"

type StatusError struct {
	Err        error
	StatusCode int
	Body       string
}

func NewStatusError(status int, body string) error {
	return &StatusError{
		Err:        errors.Errorf("received unexpected status code: %d, response=%s", status, body),
		StatusCode: status,
		Body:       body,
	}
}
func (e *StatusError) Error() string {
	return e.Err.Error()
}
