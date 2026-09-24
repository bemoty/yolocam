package yolocam

import (
	"encoding"
	"testing"
)

func checkEnum[T ~int, PT interface {
	*T
	encoding.TextUnmarshaler
}](t *testing.T, v T, wire int64, text string) {
	t.Helper()
	if got := intValue(v).GetIntValue(); got != wire {
		t.Errorf("%T %s encodes as %d, want %d", v, text, got, wire)
	}
	var parsed T
	if err := PT(&parsed).UnmarshalText([]byte(text)); err != nil || parsed != v {
		t.Errorf("%T UnmarshalText(%q) = (%d, %v), want %d", v, text, parsed, err, v)
	}
}

func TestEnumWireValues(t *testing.T) {
	checkEnum(t, ExposureAuto, 0, "auto")
	checkEnum(t, ExposureManual, 1, "manual")
	checkEnum(t, WhiteBalanceManual, 0, "manual")
	checkEnum(t, WhiteBalanceAuto, 1, "auto")
}

func TestExposureTypeUnknown(t *testing.T) {
	unknown := ExposureType(7)
	if got := unknown.String(); got != "exposure_type(7)" {
		t.Errorf("String = %q", got)
	}
	if text, err := unknown.MarshalText(); err != nil || string(text) != "7" {
		t.Errorf("MarshalText = (%q, %v), want 7", text, err)
	}
	var parsed ExposureType
	if err := parsed.UnmarshalText([]byte("7")); err == nil {
		t.Error("UnmarshalText accepted a bare number")
	}
}
