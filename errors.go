package yolocam

import (
	"errors"
	"fmt"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
	"github.com/bemoty/yolocam/internal/session"
)

var (
	ErrNotImplemented   = errors.New("property not implemented by camera")
	ErrUnsupportedValue = errors.New("unsupported value kind")
	ErrClosed           = session.ErrClosed
)

type MalformedError = session.MalformedError

type PropertyError struct {
	Op       string
	Property string
	Err      error
}

func propertyError(op string, id yolocamv1.PropertyId, err error) *PropertyError {
	return &PropertyError{Op: op, Property: propertyLabel(id), Err: err}
}

func (e *PropertyError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Property, e.Err)
}

func (e *PropertyError) Unwrap() error {
	return e.Err
}

type StatusError struct {
	Code uint32
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("camera returned status %d", e.Code)
}

type UnsupportedValueError struct {
	Kind string
	Raw  []byte
}

func (e *UnsupportedValueError) Error() string {
	return fmt.Sprintf("unsupported value kind %s (%d raw bytes)", e.Kind, len(e.Raw))
}

func (e *UnsupportedValueError) Is(target error) bool {
	return target == ErrUnsupportedValue
}

const (
	statusOK             = 200
	statusNotImplemented = 100
)

func statusError(status uint32) error {
	switch status {
	case statusOK:
		return nil
	case statusNotImplemented:
		return ErrNotImplemented
	default:
		return &StatusError{Code: status}
	}
}
