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

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

// FocusType is how the webcam focuses, named after the modes Compose offers.
type FocusType int

const (
	FocusSingle     FocusType = 1
	FocusManual     FocusType = 2
	FocusContinuous FocusType = 3
	FocusFace       FocusType = 4
)

var focusTypes = enum[FocusType]{kind: "focus_type", names: map[FocusType]string{
	FocusSingle:     "af-s",
	FocusManual:     "mf",
	FocusContinuous: "af-c",
	FocusFace:       "face",
}}

func (t FocusType) String() string {
	return focusTypes.format(t)
}

func (t FocusType) MarshalText() ([]byte, error) {
	return focusTypes.marshal(t), nil
}

func (t *FocusType) UnmarshalText(text []byte) error {
	v, err := focusTypes.unmarshal(text)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// FocusType reads the value set by [Client.SetFocusType].
func (c *Client) FocusType(ctx context.Context) (FocusType, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_FOCUS_TYPE, decodeInt[FocusType])
}

// SetFocusType switches between the webcam's autofocus modes and manual focus.
func (c *Client) SetFocusType(ctx context.Context, t FocusType) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_FOCUS_TYPE, intValue(t))
}

// ManualFocus reads the value set by [Client.SetManualFocus].
func (c *Client) ManualFocus(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_MANUAL_FOCUS, decodeInt[int])
}

// SetManualFocus sets the focus distance. Compose offers 0 to 100.
//
// The webcam only uses it while the focus type is [FocusManual].
func (c *Client) SetManualFocus(ctx context.Context, focus int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_MANUAL_FOCUS, intValue(focus))
}
