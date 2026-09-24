package camtest

import (
	"net"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/bemoty/yolocam/internal/firmware"
	"github.com/bemoty/yolocam/internal/frame"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

const Timeout = 2 * time.Second

type Camera struct {
	t      testing.TB
	Conn   net.Conn
	frames *frame.Reader
}

func Pipe(t testing.TB) (net.Conn, *Camera) {
	client, server := net.Pipe()
	return client, New(t, server)
}

func New(t testing.TB, conn net.Conn) *Camera {
	t.Cleanup(func() { _ = conn.Close() })
	return &Camera{t: t, Conn: conn, frames: frame.NewReader(conn, firmware.MaxPayload)}
}

func (c *Camera) Recv() *yolocamv1.Message {
	c.t.Helper()
	if err := c.Conn.SetReadDeadline(time.Now().Add(Timeout)); err != nil {
		c.t.Fatal(err)
	}
	payload, err := c.frames.Next()
	if err != nil {
		c.t.Fatalf("camera recv: %v", err)
	}
	msg := new(yolocamv1.Message)
	if err := proto.Unmarshal(payload, msg); err != nil {
		c.t.Fatalf("camera unmarshal: %v", err)
	}
	return msg
}

func (c *Camera) Send(msg *yolocamv1.Message) {
	c.t.Helper()
	payload, err := proto.Marshal(msg)
	if err != nil {
		c.t.Fatal(err)
	}
	c.SendRaw(payload)
}

func (c *Camera) SendRaw(payload []byte) {
	c.t.Helper()
	c.SendFrame(frame.Encode(payload, frame.DirToApp))
}

func (c *Camera) SendFrame(b []byte) {
	c.t.Helper()
	if err := c.Conn.SetWriteDeadline(time.Now().Add(Timeout)); err != nil {
		c.t.Fatal(err)
	}
	if _, err := c.Conn.Write(b); err != nil {
		c.t.Fatalf("camera send: %v", err)
	}
}

func (c *Camera) Reply(req *yolocamv1.Message, status uint32, value *yolocamv1.Value) {
	c.t.Helper()
	c.Send(&yolocamv1.Message{
		Type:     yolocamv1.MessageType_MESSAGE_TYPE_REPLY,
		Seq:      req.GetSeq(),
		Status:   status,
		Property: req.GetProperty(),
		Value:    value,
	})
}

func (c *Camera) ReplyInt(req *yolocamv1.Message, v int64) {
	c.t.Helper()
	c.Reply(req, 200, IntValue(v))
}

func (c *Camera) AcceptAttach() *yolocamv1.Message {
	c.t.Helper()
	attach := c.Recv()
	if attach.GetType() != yolocamv1.MessageType_MESSAGE_TYPE_SET ||
		attach.GetProperty() != yolocamv1.PropertyId_PROPERTY_ID_SESSION_ATTACH ||
		attach.GetValue().GetIntValue() != 1 {
		c.t.Fatalf("first frame = %v, want SET 215=1", attach)
	}
	c.Reply(attach, 200, nil)
	return attach
}

func IntValue(v int64) *yolocamv1.Value {
	return &yolocamv1.Value{Kind: &yolocamv1.Value_IntValue{IntValue: v}}
}
