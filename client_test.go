package yolocam

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protowire"

	"github.com/bemoty/yolocam/internal/camtest"
	"github.com/bemoty/yolocam/internal/firmware"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
	"github.com/bemoty/yolocam/internal/session"
)

func newTestClient(t *testing.T) (*Client, *camtest.Camera) {
	t.Helper()
	conn, cam := camtest.Pipe(t)

	type opened struct {
		s   *session.Session
		err error
	}
	ch := make(chan opened, 1)
	go func() {
		s, err := session.Open(context.Background(), conn, session.Config{MaxPayload: firmware.MaxPayload})
		ch <- opened{s, err}
	}()
	cam.AcceptAttach()
	o := <-ch
	if o.err != nil {
		t.Fatal(o.err)
	}
	c := &Client{sess: o.s, requestTimeout: camtest.Timeout}
	t.Cleanup(func() { _ = c.Close() })
	return c, cam
}

type result[T any] struct {
	v   T
	err error
}

func async[T any](fn func() (T, error)) <-chan result[T] {
	ch := make(chan result[T], 1)
	go func() {
		v, err := fn()
		ch <- result[T]{v, err}
	}()
	return ch
}

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	var r T
	select {
	case r = <-ch:
	case <-time.After(camtest.Timeout):
		t.Fatal("call did not return")
	}
	return r
}

func TestGetDecodesInt(t *testing.T) {
	c, cam := newTestClient(t)

	pending := async(func() (int, error) { return c.ExposureISO(context.Background()) })
	req := cam.Recv()
	if req.GetType() != yolocamv1.MessageType_MESSAGE_TYPE_GET || req.GetProperty() != yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO {
		t.Fatalf("request = %v, want GET exposure_iso", req)
	}
	cam.ReplyInt(req, 400)

	if r := await(t, pending); r.err != nil || r.v != 400 {
		t.Errorf("ExposureISO = (%d, %v), want 400", r.v, r.err)
	}
}

func TestSetEncodesInt(t *testing.T) {
	c, cam := newTestClient(t)

	pending := async(func() (struct{}, error) {
		return struct{}{}, c.SetExposureType(context.Background(), ExposureManual)
	})
	req := cam.Recv()
	if req.GetType() != yolocamv1.MessageType_MESSAGE_TYPE_SET || req.GetProperty() != yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE ||
		req.GetValue().GetIntValue() != int64(ExposureManual) {
		t.Fatalf("request = %v, want SET exposure_type=1", req)
	}
	cam.Reply(req, statusOK, nil)

	if r := await(t, pending); r.err != nil {
		t.Errorf("SetExposureType: %v", r.err)
	}
}

func valueWithUnknownBranch() *yolocamv1.Value {
	v := &yolocamv1.Value{}
	raw := protowire.AppendTag(nil, 77, protowire.BytesType)
	v.ProtoReflect().SetUnknown(protowire.AppendBytes(raw, []byte{1, 2}))
	return v
}

func TestGetErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  uint32
		value   *yolocamv1.Value
		wantIs  error
		wantMsg string
	}{
		{"not implemented", statusNotImplemented, nil, ErrNotImplemented, "get exposure_iso: property not implemented by camera"},
		{"unknown status", 500, nil, nil, "get exposure_iso: camera returned status 500"},
		{"missing value", statusOK, nil, ErrUnsupportedValue, "get exposure_iso: unsupported value kind none (0 raw bytes)"},
		{"wrong kind", statusOK, &yolocamv1.Value{Kind: &yolocamv1.Value_StringValue_2{StringValue_2: "x"}}, ErrUnsupportedValue, "get exposure_iso: unsupported value kind string_value_2 (3 raw bytes)"},
		{"unknown kind", statusOK, valueWithUnknownBranch(), ErrUnsupportedValue, "get exposure_iso: unsupported value kind unknown (5 raw bytes)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, cam := newTestClient(t)

			pending := async(func() (int, error) { return c.ExposureISO(context.Background()) })
			cam.Reply(cam.Recv(), tt.status, tt.value)
			err := await(t, pending).err

			if err == nil || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("err = %v, want errors.Is %v", err, tt.wantIs)
			}
		})
	}
}

func TestSetStatusError(t *testing.T) {
	c, cam := newTestClient(t)

	pending := async(func() (struct{}, error) { return struct{}{}, c.SetExposureISO(context.Background(), 800) })
	cam.Reply(cam.Recv(), statusNotImplemented, nil)
	err := await(t, pending).err

	var propErr *PropertyError
	if !errors.As(err, &propErr) || propErr.Op != "set" || !errors.Is(err, ErrNotImplemented) {
		t.Errorf("err = %v, want set *PropertyError wrapping ErrNotImplemented", err)
	}
	if got, want := err.Error(), "set exposure_iso: property not implemented by camera"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

func TestRequestTimeout(t *testing.T) {
	c, cam := newTestClient(t)
	c.requestTimeout = 20 * time.Millisecond

	pending := async(func() (int, error) { return c.ExposureISO(context.Background()) })
	cam.Recv()

	err := await(t, pending).err
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}
