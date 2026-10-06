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

package cli

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/bemoty/yolocam"
)

const (
	exitOK          = 0
	exitError       = 1
	exitConnection  = 3
	exitInterrupted = 130
)

func exitCode(err error) int {
	switch {
	case err == nil:
		return exitOK
	case errors.Is(err, context.Canceled):
		return exitInterrupted
	case isConnectionError(err):
		return exitConnection
	default:
		return exitError
	}
}

func isConnectionError(err error) bool {
	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, yolocam.ErrClosed)
}
