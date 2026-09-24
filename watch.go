package yolocam

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

type MessageType int

const (
	MessageGet   = MessageType(yolocamv1.MessageType_MESSAGE_TYPE_GET)
	MessageSet   = MessageType(yolocamv1.MessageType_MESSAGE_TYPE_SET)
	MessageReply = MessageType(yolocamv1.MessageType_MESSAGE_TYPE_REPLY)
	MessagePush  = MessageType(yolocamv1.MessageType_MESSAGE_TYPE_PUSH)
)

var messageTypes = enum[MessageType]{kind: "message_type", names: map[MessageType]string{
	MessageGet:   "get",
	MessageSet:   "set",
	MessageReply: "reply",
	MessagePush:  "push",
}}

func (t MessageType) String() string {
	return messageTypes.format(t)
}

func (t MessageType) MarshalText() ([]byte, error) {
	return messageTypes.marshal(t), nil
}

type Message struct {
	Type         MessageType
	PropertyID   int
	Seq          uint32
	Status       uint32
	DroppedSoFar uint64
	Raw          []byte
	Err          error
	msg          *yolocamv1.Message
}

func (m Message) PropertyName() string {
	name, _ := propertyName(yolocamv1.PropertyId(m.PropertyID))
	return name
}

func (m Message) ValueJSON() ([]byte, error) {
	value := m.msg.GetValue()
	if value == nil {
		return []byte("null"), nil
	}
	return protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}.Marshal(value)
}

func (c *Client) Watch(ctx context.Context, fn func(Message) error) error {
	sub := c.sess.Subscribe(watchBuffer)
	defer sub.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case f, ok := <-sub.Frames():
			if !ok {
				return c.sess.Err()
			}
			msg := Message{
				Type:         MessageType(f.Message.GetType()),
				PropertyID:   int(f.Message.GetProperty()),
				Seq:          f.Message.GetSeq(),
				Status:       f.Message.GetStatus(),
				DroppedSoFar: sub.Dropped(),
				Raw:          f.Raw,
				Err:          f.Err,
				msg:          f.Message,
			}
			if err := fn(msg); err != nil {
				return err
			}
		}
	}
}
