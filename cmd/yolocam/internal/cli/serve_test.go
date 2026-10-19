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
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/bemoty/yolocam"
	"github.com/bemoty/yolocam/cmd/yolocam/internal/jsonrpc"
	"github.com/bemoty/yolocam/internal/camtest"
)

func connectFakeCamera(t *testing.T) (*yolocam.Client, *camtest.Camera) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan *camtest.Camera, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		cam := camtest.New(t, conn)
		cam.AcceptAttach()
		accepted <- cam
	}()
	port := listener.Addr().(*net.TCPAddr).Port
	client, err := yolocam.Connect(context.Background(), &yolocam.Options{Host: "127.0.0.1", Port: port, RequestTimeout: camtest.Timeout})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client, <-accepted
}

func TestCameraMethods(t *testing.T) {
	client, cam := connectFakeCamera(t)
	handle := cameraMethods(client, new(atomic.Bool))

	tests := []struct {
		name   string
		in     string
		status uint32
		want   string
	}{
		{"set", `{"jsonrpc":"2.0","id":1,"method":"set","params":{"property":"exposure_type","value":"manual"}}`, 200,
			`{"jsonrpc":"2.0","result":null,"id":1}`},
		{"camera refuses", `{"jsonrpc":"2.0","id":2,"method":"get","params":{"property":"exposure_iso"}}`, 100,
			`{"jsonrpc":"2.0","error":{"code":1,"message":"get exposure_iso: property not implemented by camera","data":{"kind":"not_implemented","property":"exposure_iso"}},"id":2}`},
		{"unknown property", `{"jsonrpc":"2.0","id":3,"method":"get","params":{"property":"bogus"}}`, 0,
			`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params","data":{"kind":"unknown_property","property":"bogus"}},"id":3}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != 0 {
				go func() { cam.Reply(cam.Recv(), tt.status, nil) }()
			}
			var out bytes.Buffer
			if err := jsonrpc.Serve(context.Background(), strings.NewReader(tt.in+"\n"), &out, handle); err != nil {
				t.Fatal(err)
			}
			if got := strings.TrimSpace(out.String()); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestServeConnectionLost(t *testing.T) {
	client, cam := connectFakeCamera(t)
	in, _ := io.Pipe()
	done := make(chan error, 1)
	go func() { done <- serveRPC(context.Background(), client, in, io.Discard) }()

	_ = cam.Conn.Close()
	if err := <-done; !errors.Is(err, yolocam.ErrConnectionLost) {
		t.Errorf("err = %v, want ErrConnectionLost", err)
	}
}

func TestOpenRPCMethodsExist(t *testing.T) {
	data, err := openRPCDocument()
	if err != nil {
		t.Fatal(err)
	}
	var doc struct{ Methods []struct{ Name string } }
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}

	client, _ := connectFakeCamera(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	handle := cameraMethods(client, new(atomic.Bool))
	for _, m := range doc.Methods {
		_, err := handle(ctx, m.Name, nil)
		if rpcErr, ok := errors.AsType[*jsonrpc.Error](err); ok && rpcErr.Code == jsonrpc.CodeMethodNotFound {
			t.Errorf("spec lists %q but the server has no such method", m.Name)
		}
	}
}
