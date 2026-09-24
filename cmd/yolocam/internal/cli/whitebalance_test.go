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
