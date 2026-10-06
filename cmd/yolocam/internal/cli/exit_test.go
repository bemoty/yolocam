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
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/bemoty/yolocam"
)

func TestExitCode(t *testing.T) {
	dialFailure := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	inProperty := func(err error) error {
		return &yolocam.PropertyError{Op: "get", Property: "exposure_iso", Err: err}
	}

	tests := []struct {
		name string
		err  error
		want int
	}{
		{"success", nil, exitOK},
		{"bad property name", errors.New(`set: unknown property "bogus"`), exitError},
		{"camera status", inProperty(&yolocam.StatusError{Code: 500}), exitError},
		{"not implemented", inProperty(yolocam.ErrNotImplemented), exitError},
		{"dial failure", fmt.Errorf("connect 192.0.2.1: %w", dialFailure), exitConnection},
		{"request timeout", inProperty(context.DeadlineExceeded), exitConnection},
		{"camera hung up", inProperty(fmt.Errorf("read: %w", io.ErrUnexpectedEOF)), exitConnection},
		{"session closed", inProperty(yolocam.ErrClosed), exitConnection},
		{"interrupted", inProperty(context.Canceled), exitInterrupted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCode(tt.err); got != tt.want {
				t.Errorf("exitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
