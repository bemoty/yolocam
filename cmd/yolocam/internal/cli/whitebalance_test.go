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

package cli

import (
	"testing"

	"github.com/bemoty/yolocam"
)

func TestParseWhiteBalanceOffset(t *testing.T) {
	tests := []struct {
		raw    string
		want   yolocam.WhiteBalanceOffset
		wantOK bool
	}{
		{"red=10,blue=12", yolocam.WhiteBalanceOffset{Red: 10, Blue: 12}, true},
		{"blue=12,red=10", yolocam.WhiteBalanceOffset{Red: 10, Blue: 12}, true},
		{"red=0,blue=0", yolocam.WhiteBalanceOffset{}, true},
		{"red=10", yolocam.WhiteBalanceOffset{}, false},
		{"10,12", yolocam.WhiteBalanceOffset{}, false},
		{"red=10,green=12", yolocam.WhiteBalanceOffset{}, false},
		{"red=x,blue=12", yolocam.WhiteBalanceOffset{}, false},
		{"", yolocam.WhiteBalanceOffset{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := parseWhiteBalanceOffset(tt.raw)
			if (err == nil) != tt.wantOK || (tt.wantOK && got != tt.want) {
				t.Errorf("parse(%q) = (%v, %v), want (%v, ok=%v)", tt.raw, got, err, tt.want, tt.wantOK)
			}
		})
	}
}
