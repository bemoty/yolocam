package yolocam

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/bemoty/yolocam/internal/camtest"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

func TestMessageValueJSONShowsZeroFields(t *testing.T) {
	rect := &yolocamv1.Value{Kind: &yolocamv1.Value_Rect{Rect: &yolocamv1.Rect{X2: 0.5}}}
	got, err := Message{msg: &yolocamv1.Message{Value: rect}}.ValueJSON()
	if err != nil || !bytes.Equal(bytes.ReplaceAll(got, []byte(" "), nil), []byte(`{"rect":{"x1":0,"y1":0,"x2":0.5,"y2":0}}`)) {
		t.Errorf("ValueJSON = (%s, %v), want zeros shown", got, err)
	}
}

func sendUntilSubscribed(t *testing.T, cam *camtest.Camera, raw []byte, received <-chan Message) Message {
	t.Helper()
	deadline := time.After(camtest.Timeout)
	for {
		cam.SendRaw(raw)
		select {
		case p := <-received:
			return p
		case <-time.After(5 * time.Millisecond):
		case <-deadline:
			t.Fatal("Watch never subscribed")
		}
	}
}

func TestWatch(t *testing.T) {
	c, cam := newTestClient(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	received := make(chan Message, 64)
	done := make(chan error, 1)
	go func() {
		done <- c.Watch(ctx, func(m Message) error {
			received <- m
			return nil
		})
	}()

	raw, err := proto.Marshal(&yolocamv1.Message{
		Type:     yolocamv1.MessageType_MESSAGE_TYPE_PUSH,
		Seq:      1,
		Property: yolocamv1.PropertyId_PROPERTY_ID_TRACKING_RECT_PUSH,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := sendUntilSubscribed(t, cam, raw, received)
	if got.Type != MessagePush || got.PropertyID != 59 || got.Seq != 1 || !bytes.Equal(got.Raw, raw) || got.Err != nil {
		t.Errorf("message = %+v, want tracking push with raw bytes", got)
	}

	cancel()
	if err := await(t, done); !errors.Is(err, context.Canceled) {
		t.Errorf("Watch returned %v, want context.Canceled", err)
	}
}
