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
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/bemoty/yolocam/internal/frame"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

var ErrClosed = errors.New("connection closed")

type Config struct {
	MaxPayload        int
	HeartbeatInterval time.Duration
}

type response struct {
	msg *yolocamv1.Message
	err error
}

type Session struct {
	conn   net.Conn
	frames *frame.Reader
	seq    atomic.Uint32

	writeMu sync.Mutex

	mu            sync.Mutex
	pending       map[replyKey]chan response
	subscriptions map[*Subscription]struct{}
	err           error
	done          chan struct{}

	closeOnce sync.Once
	closeErr  error
	workers   sync.WaitGroup
}

func Dial(ctx context.Context, addr string, cfg Config) (*Session, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return Open(ctx, conn, cfg)
}

func Open(ctx context.Context, conn net.Conn, cfg Config) (*Session, error) {
	s := &Session{
		conn:          conn,
		frames:        frame.NewReader(conn, cfg.MaxPayload),
		pending:       make(map[replyKey]chan response),
		subscriptions: make(map[*Subscription]struct{}),
		done:          make(chan struct{}),
	}
	s.workers.Go(s.readLoop)

	if _, err := s.attach(ctx); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("attach: %w", err)
	}
	if cfg.HeartbeatInterval > 0 {
		s.workers.Go(func() { s.heartbeat(cfg.HeartbeatInterval) })
	}
	return s, nil
}

func (s *Session) Close() error {
	s.shutdown(ErrClosed)
	s.workers.Wait()
	return s.closeErr
}

func (s *Session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

func (s *Session) Request(ctx context.Context, msgType yolocamv1.MessageType, property yolocamv1.PropertyId, value *yolocamv1.Value) (*yolocamv1.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	seq := s.seq.Add(1)
	key := replyKey{seq: seq, property: property}
	reply := make(chan response, 1)
	if err := s.register(key, reply); err != nil {
		return nil, err
	}
	defer s.unregister(key)

	payload, err := proto.Marshal(&yolocamv1.Message{
		Type:     msgType,
		Seq:      seq,
		Property: property,
		Value:    value,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	if err := s.write(ctx, frame.Encode(payload, frame.DirToCamera)); err != nil {
		return nil, err
	}

	select {
	case r := <-reply:
		return r.msg, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.done:
		return nil, s.Err()
	}
}

func (s *Session) attach(ctx context.Context) (*yolocamv1.Message, error) {
	return s.Request(ctx, yolocamv1.MessageType_MESSAGE_TYPE_SET, yolocamv1.PropertyId_PROPERTY_ID_SESSION_ATTACH,
		&yolocamv1.Value{Kind: &yolocamv1.Value_IntValue{IntValue: 1}})
}

func (s *Session) register(key replyKey, reply chan response) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.pending[key] = reply
	return nil
}

func (s *Session) unregister(key replyKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pending, key)
}

func (s *Session) takePending(key replyKey) (chan response, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	reply, ok := s.pending[key]
	delete(s.pending, key)
	return reply, ok
}

func (s *Session) write(ctx context.Context, b []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	deadline, _ := ctx.Deadline()
	if err := s.conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	interrupted := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		_ = s.conn.SetWriteDeadline(time.Unix(1, 0))
		close(interrupted)
	})
	n, err := s.conn.Write(b)
	if !stop() {
		<-interrupted
	}
	if err == nil {
		return nil
	}

	err = fmt.Errorf("write request: %w", err)
	if n > 0 {
		// A partial frame leaves the camera's parser mid-frame; the stream is unusable.
		s.shutdown(err)
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}

func (s *Session) readLoop() {
	defer s.closeSubscriptions()
	for {
		payload, err := s.frames.Next()
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		if err != nil {
			s.shutdown(fmt.Errorf("read: %w", err))
			return
		}

		msg := new(yolocamv1.Message)
		if err := proto.Unmarshal(payload, msg); err != nil {
			s.publish(Frame{Message: &yolocamv1.Message{}, Raw: payload, Err: &MalformedError{Payload: bytes.Clone(payload), Err: err}})
			continue
		}
		s.dispatch(msg, payload)
	}
}

// Replies are keyed by (seq, property) and not by type: the camera answers GET 213 with a
// PUSH-typed frame, and PUSH 59 reuses seq 1 unrelated to any request.
func (s *Session) dispatch(msg *yolocamv1.Message, payload []byte) {
	if reply, ok := s.takePending(replyKey{seq: msg.GetSeq(), property: msg.GetProperty()}); ok {
		reply <- response{msg: msg}
		return
	}
	s.publish(Frame{Message: msg, Raw: payload})
}

func (s *Session) heartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), interval)
			_, err := s.attach(ctx)
			cancel()
			if err != nil {
				s.shutdown(fmt.Errorf("heartbeat: %w", err))
				return
			}
		}
	}
}

func (s *Session) shutdown(cause error) {
	s.mu.Lock()
	if s.err == nil {
		s.err = cause
		close(s.done)
	}
	s.mu.Unlock()
	s.closeOnce.Do(func() { s.closeErr = s.conn.Close() })
}
