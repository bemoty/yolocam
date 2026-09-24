// Copyright 2026 Joshua Winkler and The yolocam Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yolocam

import (
	"errors"
	"fmt"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
	"github.com/bemoty/yolocam/internal/session"
)

var (
	// ErrNotImplemented means the webcam doesn't know the property (status 100).
	ErrNotImplemented = errors.New("property not implemented by camera")
	// ErrUnsupportedValue means the webcam replied with a value this library doesn't understand. Every
	// [*UnsupportedValueError] matches it, so errors.Is(err, ErrUnsupportedValue) catches them all.
	ErrUnsupportedValue = errors.New("unsupported value kind")
	// ErrClosed means the [Client] was closed with [Client.Close].
	ErrClosed = session.ErrClosed
)

// MalformedError means the webcam sent a message this library couldn't decode at all. It only shows up in
// [Message.Err].
type MalformedError = session.MalformedError

// PropertyError is returned by every getter and setter. See the package documentation for what Err can be.
type PropertyError struct {
	Op       string // "get" or "set"
	Property string // e.g. "exposure_iso"
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

// StatusError means the webcam replied with a status this library doesn't know.
type StatusError struct {
	Code uint32
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("camera returned status %d", e.Code)
}

// UnsupportedValueError means the webcam replied with a value of a different type than this library expects for the
// property.
type UnsupportedValueError struct {
	Kind string // e.g. "int_value"; "none" if there was no value, "unknown" if the schema doesn't know it
	Raw  []byte // the value as it came off the wire, as protobuf
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
