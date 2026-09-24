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

	"google.golang.org/protobuf/encoding/protojson"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

// MessageType is the kind of a [Message] (e.g. [MessageGet], [MessageSet], ...)
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

// Message is something the webcam sent without being asked, as delivered by [Client.Watch]. That's mostly pushes,
// but also replies that arrived after their request already gave up.
type Message struct {
	Type       MessageType
	PropertyID int
	Seq        uint32
	Status     uint32
	// DroppedSoFar counts the messages this Watch skipped so far because the callback was too slow.
	DroppedSoFar uint64
	// Raw is the message as protobuf. It's set even if the message couldn't be decoded.
	Raw []byte
	// Err is a [*MalformedError] if the message couldn't be decoded.
	Err error
	msg *yolocamv1.Message
}

// PropertyName returns the name of the message's property, e.g. "tracking_rect_push". It's empty for properties the
// schema doesn't list.
func (m Message) PropertyName() string {
	name, _ := propertyName(yolocamv1.PropertyId(m.PropertyID))
	return name
}

// ValueJSON returns the message's value as JSON, or null if there is none. The keys come directly from the protobuf
// schema in proto/, so they can change whenever the schema is renamed.
func (m Message) ValueJSON() ([]byte, error) {
	value := m.msg.GetValue()
	if value == nil {
		return []byte("null"), nil
	}
	return protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}.Marshal(value)
}

// Watch calls fn for every message the webcam sends without being asked, until ctx is done, the connection ends, or
// fn returns an error. It returns whichever of those happened first.
//
// fn runs on the goroutine that called Watch. If fn takes too long, Watch skips the oldest messages instead of falling
// behind; [Message.DroppedSoFar] tells you how many.
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
