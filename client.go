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

package yolocam

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/bemoty/yolocam/internal/firmware"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
	"github.com/bemoty/yolocam/internal/session"
)

const (
	defaultRequestTimeout = 5 * time.Second
	watchBuffer           = 64
)

// Options configure how [Connect] reaches the webcam. Passing nil to [Connect] is the same as passing an empty
// Options, and any field left at its zero value falls back to its default.
type Options struct {
	Host           string        // defaults to 192.168.123.10
	Port           int           // defaults to 12345
	RequestTimeout time.Duration // per request, and for connecting; defaults to 5 seconds
}

func (o *Options) withDefaults() Options {
	var resolved Options
	if o != nil {
		resolved = *o
	}
	if resolved.Host == "" {
		resolved.Host = firmware.Host
	}
	if resolved.Port == 0 {
		resolved.Port = firmware.Port
	}
	if resolved.RequestTimeout == 0 {
		resolved.RequestTimeout = defaultRequestTimeout
	}
	return resolved
}

// Client is a connection to the webcam. It is safe to use from multiple goroutines at once.
//
// To keep the connection alive, the Client pings the webcam every 30 seconds. If the webcam stops answering, the
// Client closes itself and every call from then on returns the error that caused it.
//
// Call [Client.Close] when you're done with the webcam.
type Client struct {
	sess           *session.Session
	requestTimeout time.Duration
}

// Connect opens a connection to the webcam and returns a [Client] for it. Connecting has to finish within the
// [Options.RequestTimeout], or before ctx is done, whichever comes first.
//
// Make sure nothing else is connected to the webcam, including YoloLiv Compose. See package documentation for more info.
func Connect(ctx context.Context, opts *Options) (*Client, error) {
	o := opts.withDefaults()
	ctx, cancel := context.WithTimeout(ctx, o.RequestTimeout)
	defer cancel()

	sess, err := session.Dial(ctx, net.JoinHostPort(o.Host, strconv.Itoa(o.Port)), session.Config{
		MaxPayload:        firmware.MaxPayload,
		HeartbeatInterval: firmware.HeartbeatInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", o.Host, err)
	}
	return &Client{sess: sess, requestTimeout: o.RequestTimeout}, nil
}

// Close closes the connection to the webcam. Requests still waiting for a reply, and any running [Client.Watch]
// return [ErrClosed]. Calling Close more than once is fine.
func (c *Client) Close() error {
	return c.sess.Close()
}

func (c *Client) get[T any](ctx context.Context, id yolocamv1.PropertyId, decode func(*yolocamv1.Value) (T, error)) (T, error) {
	var value T
	reply, err := c.roundTrip(ctx, yolocamv1.MessageType_MESSAGE_TYPE_GET, id, nil)
	if err == nil {
		value, err = decode(reply.GetValue())
	}
	if err != nil {
		return value, propertyError("get", id, err)
	}
	return value, nil
}

func (c *Client) set(ctx context.Context, id yolocamv1.PropertyId, value *yolocamv1.Value) error {
	if _, err := c.roundTrip(ctx, yolocamv1.MessageType_MESSAGE_TYPE_SET, id, value); err != nil {
		return propertyError("set", id, err)
	}
	return nil
}

func (c *Client) roundTrip(ctx context.Context, msgType yolocamv1.MessageType, id yolocamv1.PropertyId, value *yolocamv1.Value) (*yolocamv1.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	reply, err := c.sess.Request(ctx, msgType, id, value)
	if err != nil {
		return nil, err
	}
	if err := statusError(reply.GetStatus()); err != nil {
		return nil, err
	}
	return reply, nil
}
