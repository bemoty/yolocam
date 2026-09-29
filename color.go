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
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

// HSLColor is one of the eight color ranges the HSL adjustment works on.
type HSLColor int

const (
	HSLRed HSLColor = iota
	HSLOrange
	HSLYellow
	HSLGreen
	HSLCyan
	HSLBlue
	HSLPurple
	HSLMagenta
	hslColorCount
)

var hslColors = enum[HSLColor]{kind: "hsl_color", names: map[HSLColor]string{
	HSLRed:     "red",
	HSLOrange:  "orange",
	HSLYellow:  "yellow",
	HSLGreen:   "green",
	HSLCyan:    "cyan",
	HSLBlue:    "blue",
	HSLPurple:  "purple",
	HSLMagenta: "magenta",
}}

func (c HSLColor) String() string {
	return hslColors.format(c)
}

func (c HSLColor) MarshalText() ([]byte, error) {
	return hslColors.marshal(c), nil
}

func (c *HSLColor) UnmarshalText(text []byte) error {
	v, err := hslColors.unmarshal(text)
	if err != nil {
		return err
	}
	*c = v
	return nil
}

const (
	MinHSLShift = -100
	MaxHSLShift = 100
)

// HSLShift moves one color range's hue, saturation, and lightness, each from [MinHSLShift] to [MaxHSLShift].
type HSLShift struct {
	Hue        int `json:"hue"`
	Saturation int `json:"saturation"`
	Lightness  int `json:"lightness"`
}

type hslChannel int

const (
	hslHue hslChannel = iota
	hslSaturation
	hslLightness
	hslChannelCount
)

func (s *HSLShift) channel(ch hslChannel) *int {
	return [...]*int{hslHue: &s.Hue, hslSaturation: &s.Saturation, hslLightness: &s.Lightness}[ch]
}

var hslChannelNames = [hslChannelCount]string{
	hslHue:        "hue",
	hslSaturation: "saturation",
	hslLightness:  "lightness",
}

// HSL holds the shift for each color range. A missing color range has no shift.
type HSL map[HSLColor]HSLShift

func (h HSL) String() string {
	var parts []string
	for _, color := range slices.Sorted(maps.Keys(h)) {
		shift := h[color]
		for ch := range hslChannelCount {
			if v := *shift.channel(ch); v != 0 {
				parts = append(parts, fmt.Sprintf("%s.%s=%d", color, hslChannelNames[ch], v))
			}
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ",")
}

// hslFieldNumber follows the wire layout: all eight hues first, then all saturations, then all lightnesses.
func hslFieldNumber(color HSLColor, ch hslChannel) protoreflect.FieldNumber {
	return protoreflect.FieldNumber(int(ch)*int(hslColorCount) + int(color) + 1)
}

func (h HSL) value() (*yolocamv1.Value, error) {
	msg := &yolocamv1.Hsl{}
	fields := msg.ProtoReflect().Descriptor().Fields()
	for color := range hslColorCount {
		shift := h[color]
		for ch := range hslChannelCount {
			v := *shift.channel(ch)
			if v < MinHSLShift || v > MaxHSLShift {
				return nil, fmt.Errorf("%s shift %d is outside %d to %d", color, v, MinHSLShift, MaxHSLShift)
			}
			fd := fields.ByNumber(hslFieldNumber(color, ch))
			msg.ProtoReflect().Set(fd, protoreflect.ValueOfString(strconv.Itoa(v)))
		}
	}
	return &yolocamv1.Value{Kind: &yolocamv1.Value_Hsl{Hsl: msg}}, nil
}

func decodeHSL(v *yolocamv1.Value) (HSL, error) {
	kind, ok := v.GetKind().(*yolocamv1.Value_Hsl)
	if !ok {
		return nil, unsupportedValue(v)
	}

	msg := kind.Hsl.ProtoReflect()
	fields := msg.Descriptor().Fields()
	h := HSL{}
	for color := range hslColorCount {
		var shift HSLShift
		for ch := range hslChannelCount {
			raw := msg.Get(fields.ByNumber(hslFieldNumber(color, ch))).String()
			if raw == "" {
				continue
			}
			n, err := strconv.Atoi(raw)
			if err != nil {
				return nil, fmt.Errorf("%w: hsl field %q is not an integer", ErrUnsupportedValue, raw)
			}
			*shift.channel(ch) = n
		}
		if shift != (HSLShift{}) {
			h[color] = shift
		}
	}
	return h, nil
}

// Saturation reads the value set by [Client.SetSaturation].
func (c *Client) Saturation(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_SATURATION, decodeInt[int])
}

// SetSaturation sets the image's color saturation. Compose offers 0 (black and white) to 100.
func (c *Client) SetSaturation(ctx context.Context, saturation int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_SATURATION, intValue(saturation))
}

// HSLEnabled reports whether the webcam applies the HSL adjustment set by [Client.SetHSL].
func (c *Client) HSLEnabled(ctx context.Context) (bool, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_HSL_ENABLED, decodeBool)
}

// SetHSLEnabled turns the HSL adjustment on or off without changing its values.
func (c *Client) SetHSLEnabled(ctx context.Context, enabled bool) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_HSL_ENABLED, boolValue(enabled))
}

// HSL reads the value set by [Client.SetHSL].
func (c *Client) HSL(ctx context.Context) (HSL, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_HSL, decodeHSL)
}

// SetHSL replaces the whole HSL adjustment, resetting every color range missing from h. Shifts outside
// [MinHSLShift] to [MaxHSLShift] are rejected before anything is sent.
//
// The webcam only uses it while [Client.HSLEnabled] is true.
func (c *Client) SetHSL(ctx context.Context, h HSL) error {
	value, err := h.value()
	if err != nil {
		return propertyError("set", yolocamv1.PropertyId_PROPERTY_ID_HSL, err)
	}
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_HSL, value)
}
