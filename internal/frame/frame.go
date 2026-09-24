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

package frame

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	magicOuter = 0x08
	magicInner = 0xA5
	terminator = 0x5A

	outerHeaderLen = 8
	innerHeaderLen = 6
	trailerLen     = 2
	headerLen      = outerHeaderLen + innerHeaderLen

	paddingOffset     = 2
	paddingLen        = 2
	restLenOffset     = paddingOffset + paddingLen
	innerMagicOffset  = outerHeaderLen
	payloadLenOffset  = innerMagicOffset + 1
	headerCheckOffset = headerLen - 1
)

type Direction byte

const (
	DirToCamera Direction = 0x01
	DirToApp    Direction = 0x00
)

var (
	ErrBadMagic        = errors.New("frame: bad magic")
	ErrPayloadTooLarge = errors.New("frame: payload too large")
	ErrDesync          = errors.New("frame: stream desynced")
)

func Encode(payload []byte, direction Direction) []byte {
	out := make([]byte, 0, frameLen(len(payload)))
	out = append(out, magicOuter, byte(direction), 0, 0)
	out = binary.LittleEndian.AppendUint32(out, uint32(restLen(len(payload))))
	out = append(out, magicInner)
	out = binary.BigEndian.AppendUint32(out, uint32(len(payload)))
	out = append(out, xorChecksum(out[outerHeaderLen:]))
	out = append(out, payload...)
	out = append(out, xorChecksum(payload), terminator)
	return out
}

type Reader struct {
	scanner *bufio.Scanner
}

func NewReader(r io.Reader, maxPayload int) *Reader {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(nil, frameLen(maxPayload))
	scanner.Split(splitter(maxPayload))
	return &Reader{scanner: scanner}
}

func (r *Reader) Next() ([]byte, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	return r.scanner.Bytes(), nil
}

func splitter(maxPayload int) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if len(data) < headerLen {
			return incomplete(atEOF)
		}

		if data[0] != magicOuter || data[innerMagicOffset] != magicInner {
			return 0, nil, fmt.Errorf("%w: outer=%#x inner=%#x", ErrBadMagic, data[0], data[innerMagicOffset])
		}
		if data[headerCheckOffset] != xorChecksum(data[outerHeaderLen:headerCheckOffset]) {
			return 0, nil, fmt.Errorf("%w: header checksum mismatch", ErrDesync)
		}

		wirePayloadLen := binary.BigEndian.Uint32(data[payloadLenOffset:headerCheckOffset])
		if uint64(wirePayloadLen) > uint64(maxPayload) {
			return 0, nil, fmt.Errorf("%w: %d bytes, max %d", ErrPayloadTooLarge, wirePayloadLen, maxPayload)
		}
		payloadLen := int(wirePayloadLen)

		wireRestLen := binary.LittleEndian.Uint32(data[restLenOffset:outerHeaderLen])
		if uint64(wireRestLen) != uint64(restLen(payloadLen)) {
			return 0, nil, fmt.Errorf("%w: outer length %d disagrees with payload length %d", ErrDesync, wireRestLen, payloadLen)
		}

		total := frameLen(payloadLen)
		if len(data) < total {
			return incomplete(atEOF)
		}

		payload := data[headerLen : headerLen+payloadLen]
		trailer := data[headerLen+payloadLen : total]
		if trailer[0] != xorChecksum(payload) || trailer[1] != terminator {
			return 0, nil, fmt.Errorf("%w: payload checksum mismatch", ErrDesync)
		}
		return total, payload, nil
	}
}

func incomplete(atEOF bool) (int, []byte, error) {
	if atEOF {
		return 0, nil, io.ErrUnexpectedEOF
	}
	return 0, nil, nil
}

func restLen(payloadLen int) int {
	return innerHeaderLen + payloadLen + trailerLen
}

func frameLen(payloadLen int) int {
	return outerHeaderLen + restLen(payloadLen)
}

func xorChecksum(b []byte) byte {
	var acc byte
	for _, v := range b {
		acc ^= v
	}
	return acc
}
