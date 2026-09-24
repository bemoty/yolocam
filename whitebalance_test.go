package yolocam

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"google.golang.org/protobuf/proto"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

func TestSetWhiteBalanceOffsetMatchesGoldenFrame(t *testing.T) {
	raw, err := os.ReadFile("testdata/frames/frames.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Frames []struct {
			Name       string `json:"name"`
			PayloadHex string `json:"payload_hex"`
		} `json:"frames"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	var want []byte
	for _, f := range fixture.Frames {
		if f.Name == "set_wb_red_blue_offset" {
			if want, err = hex.DecodeString(f.PayloadHex); err != nil {
				t.Fatal(err)
			}
		}
	}
	if want == nil {
		t.Fatal("fixture set_wb_red_blue_offset not found")
	}

	value, err := WhiteBalanceOffset{Red: 3, Blue: 7}.value()
	if err != nil {
		t.Fatal(err)
	}
	got, err := proto.Marshal(&yolocamv1.Message{
		Type:     yolocamv1.MessageType_MESSAGE_TYPE_SET,
		Seq:      839,
		Property: yolocamv1.PropertyId_PROPERTY_ID_WB_RED_BLUE_OFFSET,
		Value:    value,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("payload = %x, want %x", got, want)
	}
}

func TestDecodeWhiteBalanceOffset(t *testing.T) {
	value := &yolocamv1.Value{Kind: &yolocamv1.Value_WbRedBlueOffset{
		WbRedBlueOffset: &yolocamv1.RedBlueOffset{RedOffset: 10, BlueOffset: 20}}}
	if got, err := decodeWhiteBalanceOffset(value); err != nil || got != (WhiteBalanceOffset{Red: 10, Blue: 20}) {
		t.Errorf("got (%v, %v), want red=10,blue=20", got, err)
	}
	if _, err := decodeWhiteBalanceOffset(intValue(1)); !errors.Is(err, ErrUnsupportedValue) {
		t.Errorf("int value err = %v, want ErrUnsupportedValue", err)
	}
}

func TestSetWhiteBalanceOffsetRejectsNegative(t *testing.T) {
	c, _ := newTestClient(t)

	err := c.SetWhiteBalanceOffset(context.Background(), WhiteBalanceOffset{Red: -1, Blue: 10})
	var propErr *PropertyError
	if !errors.As(err, &propErr) || propErr.Op != "set" || propErr.Property != "wb_red_blue_offset" {
		t.Errorf("err = %v, want *PropertyError for set wb_red_blue_offset", err)
	}
}

func TestBoolWireValues(t *testing.T) {
	for b, wire := range map[bool]int64{false: 0, true: 1} {
		if got := boolValue(b).GetIntValue(); got != wire {
			t.Errorf("boolValue(%v) = %d, want %d", b, got, wire)
		}
		if got, err := decodeBool(intValue(int(wire))); err != nil || got != b {
			t.Errorf("decodeBool(%d) = (%v, %v), want %v", wire, got, err, b)
		}
	}
	if _, err := decodeBool(intValue(2)); !errors.Is(err, ErrUnsupportedValue) {
		t.Errorf("decodeBool(2) err = %v, want ErrUnsupportedValue", err)
	}
}
