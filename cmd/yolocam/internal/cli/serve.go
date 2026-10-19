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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
	"github.com/bemoty/yolocam/cmd/yolocam/internal/jsonrpc"
)

const rpcProtocolVersion = 1

const codeCameraError = 1

type initializeResult struct {
	ProtocolVersion  int                `json:"protocol_version"`
	CLI              buildInfo          `json:"cli"`
	Device           yolocam.DeviceInfo `json:"device"`
	FirmwareVerified bool               `json:"firmware_verified"`
}

type errorData struct {
	Kind      string `json:"kind"`
	Property  string `json:"property,omitzero"`
	Reason    string `json:"reason,omitzero"`
	Status    uint32 `json:"status,omitzero"`
	ValueKind string `json:"value_kind,omitzero"`
}

func newServeCommand(opts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve --stdio",
		Short: "Hold one camera session and answer JSON-RPC 2.0 requests, one per line",
		Long:  "Hold one camera session and answer JSON-RPC 2.0 requests from stdin. See `yolocam openrpc` for the methods.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.withCamera(cmd, func(ctx context.Context, c *yolocam.Client) error {
				return serveRPC(ctx, c, cmd.InOrStdin(), cmd.OutOrStdout())
			})
		},
	}
	cmd.Flags().Bool("stdio", false, "read requests from stdin and write responses to stdout")
	_ = cmd.MarkFlagRequired("stdio")
	return cmd
}

func serveRPC(ctx context.Context, c *yolocam.Client, in io.Reader, out io.Writer) error {
	ctx, stop := context.WithCancelCause(ctx)
	out = &lockedWriter{w: out}

	var subscribed atomic.Bool
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		stop(c.Watch(ctx, func(m yolocam.Message) error {
			if !subscribed.Load() || m.Type != yolocam.MessagePush {
				return nil
			}
			return jsonrpc.Notify(out, "push", newMessageLine(m, false))
		}))
	}()
	defer func() { <-watchDone }()
	defer stop(nil)

	err := jsonrpc.Serve(ctx, in, out, cameraMethods(c, &subscribed))
	if !errors.Is(err, context.Canceled) {
		return err
	}
	if cause := context.Cause(ctx); !errors.Is(cause, context.Canceled) {
		return cause
	}
	return nil
}

type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

func cameraMethods(c *yolocam.Client, subscribed *atomic.Bool) jsonrpc.Handler {
	return func(ctx context.Context, method string, params jsontext.Value) (any, error) {
		switch method {
		case "initialize":
			info, err := c.Info(ctx)
			if err != nil {
				return nil, cameraError("", err)
			}
			return initializeResult{rpcProtocolVersion, build, info, info.FirmwareVerified()}, nil
		case "list":
			return listedProperties(), nil
		case "subscribe":
			subscribed.Store(true)
			return nil, nil
		case "unsubscribe":
			subscribed.Store(false)
			return nil, nil
		case "get":
			var p struct {
				Property string `json:"property"`
			}
			if err := decodeParams(params, &p); err != nil {
				return nil, err
			}
			prop, err := findProperty(p.Property)
			if err != nil {
				return nil, err
			}
			value, err := prop.get(ctx, c)
			if err != nil {
				return nil, cameraError(prop.Name, err)
			}
			return propertyValue{Property: prop.Name, Value: value}, nil
		case "set":
			var p struct {
				Property string         `json:"property"`
				Value    jsontext.Value `json:"value"`
			}
			if err := decodeParams(params, &p); err != nil {
				return nil, err
			}
			prop, err := findProperty(p.Property)
			if err != nil {
				return nil, err
			}
			apply, err := prop.parseJSONValue(p.Value)
			if err != nil {
				return nil, jsonrpc.InvalidParams(errorData{Kind: "invalid_value", Property: prop.Name, Reason: err.Error()})
			}
			if err := apply(ctx, c); err != nil {
				return nil, cameraError(prop.Name, err)
			}
			return nil, nil
		}
		return nil, jsonrpc.MethodNotFound()
	}
}

func decodeParams(params jsontext.Value, v any) error {
	if params.Kind() != '{' {
		return jsonrpc.InvalidParams(errorData{Kind: "invalid_params", Reason: "params must be an object"})
	}
	if err := json.Unmarshal(params, v, json.RejectUnknownMembers(true)); err != nil {
		return jsonrpc.InvalidParams(errorData{Kind: "invalid_params", Reason: err.Error()})
	}
	return nil
}

func findProperty(name string) (property, error) {
	p, ok := lookupProperty(name)
	if !ok {
		return property{}, jsonrpc.InvalidParams(errorData{Kind: "unknown_property", Property: name})
	}
	return p, nil
}

func cameraError(property string, err error) error {
	data := errorData{Property: property}
	switch {
	case errors.Is(err, yolocam.ErrInvalidValue):
		return jsonrpc.InvalidParams(errorData{Kind: "invalid_value", Property: property, Reason: err.Error()})
	case errors.Is(err, yolocam.ErrNotImplemented):
		data.Kind = "not_implemented"
	case errors.Is(err, yolocam.ErrUnsupportedValue):
		data.Kind = "unsupported_value"
		if unsupported, ok := errors.AsType[*yolocam.UnsupportedValueError](err); ok {
			data.ValueKind = unsupported.Kind
		}
	case errors.Is(err, yolocam.ErrConnectionLost), errors.Is(err, yolocam.ErrClosed):
		data = errorData{Kind: "session_lost"}
	case errors.Is(err, context.DeadlineExceeded):
		data.Kind = "timeout"
	default:
		status, ok := errors.AsType[*yolocam.StatusError](err)
		if !ok {
			return err
		}
		data.Kind = "camera_status"
		data.Status = status.Code
	}
	return &jsonrpc.Error{Code: codeCameraError, Message: err.Error(), Data: data}
}
