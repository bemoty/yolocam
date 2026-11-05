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
	"testing"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

func TestOffsetEncodedWireValues(t *testing.T) {
	tests := []struct {
		name   string
		encode func(int) *yolocamv1.Value
		decode func(*yolocamv1.Value) (int, error)
		value  int
		wire   int
	}{
		{"sharpness", sharpnessValue, decodeSharpness, -10, 70},
		{"sharpness", sharpnessValue, decodeSharpness, -1, 61},
		{"sharpness", sharpnessValue, decodeSharpness, 0, 0},
		{"sharpness", sharpnessValue, decodeSharpness, 10, 10},
		{"ev", evValue, decodeEV, -8, 0},
		{"ev", evValue, decodeEV, 0, 8},
		{"ev", evValue, decodeEV, 8, 16},
	}
	for _, tt := range tests {
		if got := tt.encode(tt.value).GetIntValue(); got != int64(tt.wire) {
			t.Errorf("%s %d encodes as %d, want %d", tt.name, tt.value, got, tt.wire)
		}
		if got, err := tt.decode(intValue(tt.wire)); err != nil || got != tt.value {
			t.Errorf("%s wire %d decodes as (%d, %v), want %d", tt.name, tt.wire, got, err, tt.value)
		}
	}
}
