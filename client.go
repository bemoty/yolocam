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
	defaultPort           = 12345
	defaultRequestTimeout = 5 * time.Second
	watchBuffer           = 64
)

type Options struct {
	Port           int
	RequestTimeout time.Duration
}

func (o *Options) withDefaults() Options {
	var resolved Options
	if o != nil {
		resolved = *o
	}
	if resolved.Port == 0 {
		resolved.Port = defaultPort
	}
	if resolved.RequestTimeout == 0 {
		resolved.RequestTimeout = defaultRequestTimeout
	}
	return resolved
}

type Client struct {
	sess           *session.Session
	requestTimeout time.Duration
}

// Connect may only hold one connection open at a time. A second one (e.g., alongside Compose) kills the wire
// until the camera is power-cycled. I haven't found a way around or to recover from this yet.
func Connect(ctx context.Context, host string, opts *Options) (*Client, error) {
	o := opts.withDefaults()
	ctx, cancel := context.WithTimeout(ctx, o.RequestTimeout)
	defer cancel()

	sess, err := session.Dial(ctx, net.JoinHostPort(host, strconv.Itoa(o.Port)), session.Config{
		MaxPayload:        firmware.MaxPayload,
		HeartbeatInterval: firmware.HeartbeatInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", host, err)
	}
	return &Client{sess: sess, requestTimeout: o.RequestTimeout}, nil
}

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
