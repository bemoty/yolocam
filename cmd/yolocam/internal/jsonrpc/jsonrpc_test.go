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

package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"errors"
	"io"
	"strings"
	"testing"
)

func echo(_ context.Context, method string, _ jsontext.Value) (any, error) {
	if method != "echo" {
		return nil, MethodNotFound()
	}
	return "ok", nil
}

func TestServe(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"notification", `{"jsonrpc":"2.0","method":"echo"}`, ``},
		{"unknown method", `{"jsonrpc":"2.0","method":"nope","id":1}`,
			`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":1}`},
		{"parse error", `{oops`,
			`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`},
		{"batch", `[{"jsonrpc":"2.0","method":"echo","id":1},{"jsonrpc":"2.0","method":"echo"},1]`,
			`[{"jsonrpc":"2.0","result":"ok","id":1},{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := Serve(context.Background(), strings.NewReader(tt.in+"\n"), &out, echo); err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(out.String()); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestServeCancel(t *testing.T) {
	in, _ := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, in, io.Discard, echo) }()

	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
