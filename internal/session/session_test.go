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

package session

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/bemoty/yolocam/internal/camtest"
	"github.com/bemoty/yolocam/internal/firmware"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

const (
	propISO      = yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO
	propTracking = yolocamv1.PropertyId_PROPERTY_ID_TRACKING_RECT_PUSH
	prop213      = yolocamv1.PropertyId_PROPERTY_ID_UNNAMED_213
	propAttach   = yolocamv1.PropertyId_PROPERTY_ID_SESSION_ATTACH

	typeGet  = yolocamv1.MessageType_MESSAGE_TYPE_GET
	typeSet  = yolocamv1.MessageType_MESSAGE_TYPE_SET
	typePush = yolocamv1.MessageType_MESSAGE_TYPE_PUSH
)

type result struct {
	msg *yolocamv1.Message
	err error
}

func request(ctx context.Context, s *Session, msgType yolocamv1.MessageType, property yolocamv1.PropertyId) <-chan result {
	out := make(chan result, 1)
	go func() {
		msg, err := s.Request(ctx, msgType, property, nil)
		out <- result{msg, err}
	}()
	return out
}

func await(t *testing.T, ch <-chan result) result {
	t.Helper()
	var r result
	select {
	case r = <-ch:
	case <-time.After(camtest.Timeout):
		t.Fatal("request did not return")
	}
	return r
}

func openSession(t *testing.T, cfg Config) (*Session, *camtest.Camera) {
	t.Helper()
	if cfg.MaxPayload == 0 {
		cfg.MaxPayload = firmware.MaxPayload
	}
	conn, cam := camtest.Pipe(t)

	type opened struct {
		s   *Session
		err error
	}
	ch := make(chan opened, 1)
	go func() {
		s, err := Open(context.Background(), conn, cfg)
		ch <- opened{s, err}
	}()

	if attach := cam.AcceptAttach(); attach.GetSeq() != 1 {
		t.Fatalf("attach seq = %d, want 1", attach.GetSeq())
	}
	o := <-ch
	if o.err != nil {
		t.Fatal(o.err)
	}
	t.Cleanup(func() { _ = o.s.Close() })
	return o.s, cam
}

func TestRequestCorrelatesOutOfOrderReplies(t *testing.T) {
	s, cam := openSession(t, Config{})
	ctx := context.Background()

	first := request(ctx, s, typeGet, propISO)
	req1 := cam.Recv()
	second := request(ctx, s, typeGet, propISO)
	req2 := cam.Recv()

	cam.ReplyInt(req2, 200)
	cam.ReplyInt(req1, 100)

	if r := await(t, first); r.err != nil || r.msg.GetValue().GetIntValue() != 100 {
		t.Errorf("first = (%v, %v), want value 100", r.msg, r.err)
	}
	if r := await(t, second); r.err != nil || r.msg.GetValue().GetIntValue() != 200 {
		t.Errorf("second = (%v, %v), want value 200", r.msg, r.err)
	}
}

func TestReplyWithPushType(t *testing.T) {
	s, cam := openSession(t, Config{})

	pending := request(context.Background(), s, typeGet, prop213)
	req := cam.Recv()
	cam.Send(&yolocamv1.Message{Type: typePush, Seq: req.GetSeq(), Status: 200, Property: prop213})

	if r := await(t, pending); r.err != nil || r.msg.GetProperty() != prop213 {
		t.Errorf("got (%v, %v), want the PUSH-typed reply to 213", r.msg, r.err)
	}
}

func TestPushBeforeReply(t *testing.T) {
	s, cam := openSession(t, Config{})
	sub := s.Subscribe(1)

	pending := request(context.Background(), s, typeGet, propISO)
	req := cam.Recv()
	push := &yolocamv1.Message{Type: typePush, Seq: 1, Property: propTracking}
	cam.Send(push)
	cam.ReplyInt(req, 400)

	if r := await(t, pending); r.err != nil || r.msg.GetValue().GetIntValue() != 400 {
		t.Errorf("got (%v, %v), want value 400", r.msg, r.err)
	}
	select {
	case f := <-sub.Frames():
		want, _ := proto.Marshal(push)
		if f.Message.GetProperty() != propTracking || f.Err != nil || !bytes.Equal(f.Raw, want) {
			t.Errorf("frame = (%v, %x, %v), want tracking push with raw %x", f.Message, f.Raw, f.Err, want)
		}
	case <-time.After(camtest.Timeout):
		t.Fatal("push not delivered")
	}
}

func TestCancelledRequestKeepsSessionUsable(t *testing.T) {
	s, cam := openSession(t, Config{})
	sub := s.Subscribe(2)

	ctx, cancel := context.WithCancel(context.Background())
	pending := request(ctx, s, typeGet, propISO)
	stale := cam.Recv()
	cancel()
	if r := await(t, pending); !errors.Is(r.err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", r.err)
	}

	next := request(context.Background(), s, typeGet, propISO)
	req := cam.Recv()
	cam.ReplyInt(stale, 100)
	cam.ReplyInt(req, 800)

	if r := await(t, next); r.err != nil || r.msg.GetValue().GetIntValue() != 800 {
		t.Errorf("got (%v, %v), want value 800", r.msg, r.err)
	}
	if len(sub.Frames()) != 1 {
		t.Fatalf("%d frames published, want only the stale reply", len(sub.Frames()))
	}
	if f := <-sub.Frames(); f.Message.GetSeq() != stale.GetSeq() || f.Message.GetValue().GetIntValue() != 100 {
		t.Errorf("published frame = %v, want the stale reply", f.Message)
	}
}

func TestReadErrorFailsPendingAndLaterRequests(t *testing.T) {
	s, cam := openSession(t, Config{})

	pending := request(context.Background(), s, typeGet, propISO)
	cam.Recv()
	_ = cam.Conn.Close()

	if r := await(t, pending); !errors.Is(r.err, io.ErrUnexpectedEOF) {
		t.Fatalf("pending err = %v, want io.ErrUnexpectedEOF", r.err)
	}
	if _, err := s.Request(context.Background(), typeGet, propISO, nil); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("later err = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestClose(t *testing.T) {
	s, _ := openSession(t, Config{})
	sub := s.Subscribe(1)

	if err := s.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
	if _, err := s.Request(context.Background(), typeGet, propISO, nil); !errors.Is(err, ErrClosed) {
		t.Errorf("Request after Close: err = %v, want ErrClosed", err)
	}
	if _, ok := <-sub.Frames(); ok {
		t.Error("subscription not closed by session Close")
	}
	if _, ok := <-s.Subscribe(1).Frames(); ok {
		t.Error("Subscribe after Close returned an open subscription")
	}
}

func TestSubscriptionDropsOldest(t *testing.T) {
	s, cam := openSession(t, Config{})
	sub := s.Subscribe(1)

	for seq := range uint32(3) {
		cam.Send(&yolocamv1.Message{Type: typePush, Seq: seq, Property: propTracking})
	}
	pending := request(context.Background(), s, typeGet, propISO)
	cam.ReplyInt(cam.Recv(), 0)
	await(t, pending)

	kept := <-sub.Frames()
	if kept.Message.GetSeq() != 2 {
		t.Errorf("kept push seq = %d, want newest (2)", kept.Message.GetSeq())
	}
	if got := sub.Dropped(); got != 2 {
		t.Errorf("Dropped = %d, want 2", got)
	}
}

func TestHeartbeat(t *testing.T) {
	_, cam := openSession(t, Config{HeartbeatInterval: 10 * time.Millisecond})

	for range 2 {
		msg := cam.Recv()
		if msg.GetType() != typeSet || msg.GetProperty() != propAttach {
			t.Fatalf("heartbeat frame = %v, want SET 215", msg)
		}
		cam.ReplyInt(msg, 0)
	}
}

func TestMissedHeartbeatFailsSession(t *testing.T) {
	s, cam := openSession(t, Config{HeartbeatInterval: 10 * time.Millisecond})

	cam.Recv()
	select {
	case <-s.done:
	case <-time.After(camtest.Timeout):
		t.Fatal("session survived an unanswered heartbeat")
	}
	if err := s.Err(); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestMalformedFrameIsPublished(t *testing.T) {
	s, cam := openSession(t, Config{})
	sub := s.Subscribe(1)

	cam.SendRaw([]byte{0x00})
	pending := request(context.Background(), s, typeGet, propISO)
	cam.ReplyInt(cam.Recv(), 0)
	await(t, pending)

	var malformed *MalformedError
	if f := <-sub.Frames(); !errors.As(f.Err, &malformed) || !bytes.Equal(f.Raw, []byte{0x00}) {
		t.Errorf("frame = (%v, %x, %v), want malformed with raw bytes", f.Message, f.Raw, f.Err)
	}
}
