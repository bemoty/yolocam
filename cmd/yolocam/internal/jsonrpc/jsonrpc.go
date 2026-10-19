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
	"bufio"
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
)

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitzero"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.Code)
}

func MethodNotFound() *Error {
	return &Error{Code: CodeMethodNotFound, Message: "Method not found"}
}

func InvalidParams(data any) *Error {
	return &Error{Code: CodeInvalidParams, Message: "Invalid params", Data: data}
}

type Handler func(ctx context.Context, method string, params jsontext.Value) (result any, err error)

type request struct {
	method   string
	params   jsontext.Value
	id       jsontext.Value
	isNotice bool
}

type response struct {
	JSONRPC string         `json:"jsonrpc"`
	Result  jsontext.Value `json:"result,omitzero"`
	Error   *Error         `json:"error,omitzero"`
	ID      jsontext.Value `json:"id"`
}

var null = jsontext.Value("null")

func Serve(ctx context.Context, in io.Reader, out io.Writer, handle Handler) error {
	lines := make(chan []byte)
	var readErr error
	go func() {
		defer close(lines)
		reader := bufio.NewReader(in)
		for {
			line, err := reader.ReadBytes('\n')
			if line = bytes.TrimSpace(line); len(line) > 0 {
				select {
				case lines <- line:
				case <-ctx.Done():
					return
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					readErr = err
				}
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case line, ok := <-lines:
			if !ok {
				return readErr
			}
			if reply := serveLine(context.WithoutCancel(ctx), line, handle); reply != nil {
				if err := writeLine(out, reply); err != nil {
					return err
				}
			}
		}
	}
}

func Notify(out io.Writer, method string, params any) error {
	return writeLine(out, struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{"2.0", method, params})
}

func writeLine(out io.Writer, v any) error {
	line, err := json.Marshal(v, json.Deterministic(true))
	if err != nil {
		return err
	}
	_, err = out.Write(append(line, '\n'))
	return err
}

func serveLine(ctx context.Context, line []byte, handle Handler) any {
	message := jsontext.Value(line)
	if !message.IsValid() {
		return errorResponse(null, &Error{Code: CodeParseError, Message: "Parse error"})
	}
	if message.Kind() != '[' {
		if reply := serveRequest(ctx, message, handle); reply != nil {
			return reply
		}
		return nil
	}

	var batch []jsontext.Value
	if err := json.Unmarshal(message, &batch); err != nil || len(batch) == 0 {
		return errorResponse(null, invalidRequest())
	}
	var replies []*response
	for _, element := range batch {
		if reply := serveRequest(ctx, element, handle); reply != nil {
			replies = append(replies, reply)
		}
	}
	if len(replies) == 0 {
		return nil
	}
	return replies
}

func serveRequest(ctx context.Context, raw jsontext.Value, handle Handler) *response {
	req, invalid := parseRequest(raw)
	if invalid != nil {
		return invalid
	}

	result, err := handle(ctx, req.method, req.params)
	if req.isNotice {
		return nil
	}
	if err != nil {
		rpcErr, ok := errors.AsType[*Error](err)
		if !ok {
			rpcErr = &Error{Code: CodeInternalError, Message: "Internal error", Data: err.Error()}
		}
		return errorResponse(req.id, rpcErr)
	}
	encoded, err := json.Marshal(result, json.Deterministic(true))
	if err != nil {
		return errorResponse(req.id, &Error{Code: CodeInternalError, Message: "Internal error", Data: err.Error()})
	}
	return &response{JSONRPC: "2.0", Result: encoded, ID: req.id}
}

func parseRequest(raw jsontext.Value) (request, *response) {
	var members map[string]jsontext.Value
	if err := json.Unmarshal(raw, &members); err != nil {
		return request{}, errorResponse(null, invalidRequest())
	}

	req := request{id: null}
	switch id, ok := members["id"]; {
	case !ok:
		req.isNotice = true
	case id.Kind() == '"' || id.Kind() == '0' || id.Kind() == 'n':
		req.id = id
	default:
		return request{}, errorResponse(null, invalidRequest())
	}

	if string(members["jsonrpc"]) != `"2.0"` || members["method"].Kind() != '"' {
		return request{}, errorResponse(req.id, invalidRequest())
	}
	if err := json.Unmarshal(members["method"], &req.method); err != nil {
		return request{}, errorResponse(req.id, invalidRequest())
	}
	if params, ok := members["params"]; ok {
		if params.Kind() != '[' && params.Kind() != '{' {
			return request{}, errorResponse(req.id, invalidRequest())
		}
		req.params = params
	}
	return req, nil
}

func invalidRequest() *Error {
	return &Error{Code: CodeInvalidRequest, Message: "Invalid Request"}
}

func errorResponse(id jsontext.Value, err *Error) *response {
	return &response{JSONRPC: "2.0", Error: err, ID: id}
}
