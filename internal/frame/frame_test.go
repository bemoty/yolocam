package frame

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/bemoty/yolocam/internal/firmware"
)

type goldenFrame struct {
	Name      string
	Direction Direction
	Frame     []byte
	Payload   []byte
}

type goldenStream struct {
	Name           string
	Chunks         [][]byte
	ExpectedFrames int
}

func loadGoldenFrames(t testing.TB) []goldenFrame {
	t.Helper()
	var fixture struct {
		Frames []struct {
			Name       string `json:"name"`
			Direction  string `json:"direction"`
			FrameHex   string `json:"frame_hex"`
			PayloadHex string `json:"payload_hex"`
		} `json:"frames"`
	}
	readFixture(t, "frames.json", &fixture)

	frames := make([]goldenFrame, 0, len(fixture.Frames))
	for _, f := range fixture.Frames {
		frames = append(frames, goldenFrame{
			Name:      f.Name,
			Direction: parseDirection(t, f.Direction),
			Frame:     decodeHex(t, f.FrameHex),
			Payload:   decodeHex(t, f.PayloadHex),
		})
	}
	return frames
}

func loadGoldenStreams(t testing.TB) []goldenStream {
	t.Helper()
	var fixture struct {
		Streams []struct {
			Name           string   `json:"name"`
			ChunksHex      []string `json:"chunks_hex"`
			ExpectedFrames int      `json:"expected_frames"`
		} `json:"streams"`
	}
	readFixture(t, "streams.json", &fixture)

	streams := make([]goldenStream, 0, len(fixture.Streams))
	for _, s := range fixture.Streams {
		chunks := make([][]byte, 0, len(s.ChunksHex))
		for _, c := range s.ChunksHex {
			chunks = append(chunks, decodeHex(t, c))
		}
		streams = append(streams, goldenStream{Name: s.Name, Chunks: chunks, ExpectedFrames: s.ExpectedFrames})
	}
	return streams
}

func readFixture(t testing.TB, name string, v any) {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/frames/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

func decodeHex(t testing.TB, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func parseDirection(t testing.TB, s string) Direction {
	t.Helper()
	switch s {
	case "to_camera":
		return DirToCamera
	case "from_camera":
		return DirToApp
	default:
		t.Fatalf("unknown fixture direction %q", s)
		return 0
	}
}

func TestEncodeGolden(t *testing.T) {
	for _, f := range loadGoldenFrames(t) {
		t.Run(f.Name, func(t *testing.T) {
			if got := Encode(f.Payload, f.Direction); !bytes.Equal(got, f.Frame) {
				t.Errorf("Encode:\n got %x\nwant %x", got, f.Frame)
			}
		})
	}
}

func TestSplitterGolden(t *testing.T) {
	split := splitter(firmware.MaxPayload)
	for _, f := range loadGoldenFrames(t) {
		t.Run(f.Name, func(t *testing.T) {
			advance, token, err := split(f.Frame, false)
			if err != nil {
				t.Fatal(err)
			}
			if advance != len(f.Frame) {
				t.Errorf("advance = %d, want %d", advance, len(f.Frame))
			}
			if !bytes.Equal(token, f.Payload) {
				t.Errorf("token:\n got %x\nwant %x", token, f.Payload)
			}
		})
	}
}

func TestReaderStreams(t *testing.T) {
	for _, s := range loadGoldenStreams(t) {
		t.Run(s.Name, func(t *testing.T) {
			readers := make([]io.Reader, 0, len(s.Chunks))
			for _, c := range s.Chunks {
				readers = append(readers, bytes.NewReader(c))
			}
			r := NewReader(io.MultiReader(readers...), firmware.MaxPayload)

			frames := 0
			for {
				_, err := r.Next()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				frames++
			}
			if frames != s.ExpectedFrames {
				t.Errorf("frames = %d, want %d", frames, s.ExpectedFrames)
			}
		})
	}
}

func TestReaderAcceptsPayloadAtLimit(t *testing.T) {
	const maxPayload = 5000
	atLimit := bytes.Repeat([]byte{0x42}, maxPayload)

	payload, err := NewReader(bytes.NewReader(Encode(atLimit, DirToApp)), maxPayload).Next()
	if err != nil {
		t.Fatalf("payload at limit: %v", err)
	}
	if !bytes.Equal(payload, atLimit) {
		t.Errorf("payload at limit came back altered")
	}
}

func TestSplitterIncomplete(t *testing.T) {
	frame := Encode([]byte{1, 2, 3}, DirToCamera)
	split := splitter(firmware.MaxPayload)

	for _, n := range []int{0, 1, headerLen - 1, headerLen, len(frame) - 1} {
		advance, token, err := split(frame[:n], false)
		if advance != 0 || token != nil || err != nil {
			t.Errorf("len %d, not EOF: got (%d, %x, %v), want request for more data", n, advance, token, err)
		}
	}

	if advance, token, err := split(nil, true); advance != 0 || token != nil || err != nil {
		t.Errorf("empty at EOF: got (%d, %x, %v), want clean end", advance, token, err)
	}
	for _, n := range []int{1, headerLen - 1, headerLen, len(frame) - 1} {
		if _, _, err := split(frame[:n], true); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("len %d at EOF: err = %v, want io.ErrUnexpectedEOF", n, err)
		}
	}
}

func TestSplitterRejects(t *testing.T) {
	valid := Encode([]byte{1, 2, 3}, DirToCamera)
	corrupt := func(offset int, value byte) []byte {
		b := bytes.Clone(valid)
		b[offset] = value
		return b
	}

	tests := []struct {
		name       string
		data       []byte
		maxPayload int
		want       error
	}{
		{"outer magic", corrupt(0, 0x09), firmware.MaxPayload, ErrBadMagic},
		{"inner magic", corrupt(innerMagicOffset, 0xA6), firmware.MaxPayload, ErrBadMagic},
		{"header checksum", corrupt(headerCheckOffset, valid[headerCheckOffset]^0xFF), firmware.MaxPayload, ErrDesync},
		{"outer length", corrupt(restLenOffset, valid[restLenOffset]+1), firmware.MaxPayload, ErrDesync},
		{"payload checksum", corrupt(len(valid)-2, valid[len(valid)-2]^0xFF), firmware.MaxPayload, ErrDesync},
		{"terminator", corrupt(len(valid)-1, 0x00), firmware.MaxPayload, ErrDesync},
		{"payload over max", valid, 2, ErrPayloadTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := splitter(tt.maxPayload)(tt.data, false); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func FuzzSplitter(f *testing.F) {
	for _, g := range loadGoldenFrames(f) {
		f.Add(g.Frame, false)
	}
	for _, s := range loadGoldenStreams(f) {
		f.Add(bytes.Join(s.Chunks, nil), true)
	}
	split := splitter(firmware.MaxPayload)

	f.Fuzz(func(t *testing.T, data []byte, atEOF bool) {
		advance, token, err := split(data, atEOF)
		if err != nil || advance == 0 {
			return
		}
		if advance > len(data) {
			t.Fatalf("advance %d exceeds input %d", advance, len(data))
		}

		want := bytes.Clone(data[:advance])
		clear(want[paddingOffset : paddingOffset+paddingLen])
		if got := Encode(token, Direction(data[1])); !bytes.Equal(got, want) {
			t.Fatalf("accepted frame does not re-encode:\n got %x\nwant %x", got, want)
		}
	})
}
