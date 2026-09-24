package session

import (
	"bytes"
	"sync/atomic"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

type Frame struct {
	Message *yolocamv1.Message
	Raw     []byte
	Err     error
}

type Subscription struct {
	session *Session
	frames  chan Frame
	dropped atomic.Uint64
}

func (s *Session) Subscribe(buffer int) *Subscription {
	sub := &Subscription{session: s, frames: make(chan Frame, max(buffer, 1))}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		close(sub.frames)
		return sub
	}
	s.subscriptions[sub] = struct{}{}
	return sub
}

func (sub *Subscription) Frames() <-chan Frame {
	return sub.frames
}

func (sub *Subscription) Dropped() uint64 {
	return sub.dropped.Load()
}

func (sub *Subscription) Close() {
	s := sub.session
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.subscriptions[sub]; ok {
		delete(s.subscriptions, sub)
		close(sub.frames)
	}
}

func (sub *Subscription) deliverDroppingOldest(f Frame) {
	for {
		select {
		case sub.frames <- f:
			return
		default:
		}
		select {
		case <-sub.frames:
			sub.dropped.Add(1)
		default:
		}
	}
}

func (s *Session) publish(f Frame) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.subscriptions) == 0 {
		return
	}
	f.Raw = bytes.Clone(f.Raw)
	for sub := range s.subscriptions {
		sub.deliverDroppingOldest(f)
	}
}

func (s *Session) closeSubscriptions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sub := range s.subscriptions {
		close(sub.frames)
	}
	clear(s.subscriptions)
}
