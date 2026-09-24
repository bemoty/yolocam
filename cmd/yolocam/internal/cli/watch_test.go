package cli

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/bemoty/yolocam"
)

func TestMessageLine(t *testing.T) {
	clean := yolocam.Message{Type: yolocam.MessagePush, PropertyID: 59, Seq: 1, Raw: []byte{0x08, 0x05}}
	malformed := yolocam.Message{PropertyID: 59, Raw: []byte{0x00}, Err: errors.New("bad wire data")}

	tests := []struct {
		name      string
		message   yolocam.Message
		alwaysRaw bool
		want      string
	}{
		{"clean omits raw", clean, false,
			`{"type":"push","property_id":59,"property":"tracking_rect_push","seq":1,"status":0,"dropped_so_far":0,"value":null}`},
		{"--raw adds raw", clean, true,
			`{"type":"push","property_id":59,"property":"tracking_rect_push","seq":1,"status":0,"dropped_so_far":0,"value":null,"raw":"0805"}`},
		{"malformed forces raw", malformed, false,
			`{"type":"0","property_id":59,"property":"tracking_rect_push","seq":0,"status":0,"dropped_so_far":0,"raw":"00","error":"bad wire data"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(newMessageLine(tt.message, tt.alwaysRaw))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}
