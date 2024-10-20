package publisher

import (
	"errors"
	"fmt"
)

// APIError is an error wrapper with API Name.
type APIError struct {
	Name string
	Err  error
}

// WrapAPIErr returns APIError if err is not nil, otherwise returns nil.
func wrapAPIErr(name string, err error) error {
	if err != nil {
		return &APIError{Name: name, Err: err}
	} else {
		return nil
	}
}

// Implement error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("%v, error = %v", e.Name, e.Err.Error())
}

// Implement error interface.
func (e *APIError) Is(err error) bool  { return errors.Is(e.Err, err) }
func (e *APIError) As(target any) bool { return errors.As(e.Err, target) }
func (e *APIError) Unwrap() error      { return e.Err }
