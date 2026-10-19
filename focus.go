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

var FocusTypeProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_FOCUS_TYPE),
	Summary: "autofocus mode, or manual focus",
	Value:   focusTypes.schema(),
}

var ManualFocusProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_MANUAL_FOCUS),
	Summary:       "manual focus distance",
	Value:         Schema{Type: "integer", Minimum: new(0.0), Maximum: new(100.0)},
	EffectiveWhen: Condition{Property: FocusTypeProperty.Name, Equals: FocusManual.String()},
}

// FocusType reads the value set by [Client.SetFocusType].
func (c *Client) FocusType(ctx context.Context) (FocusType, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_FOCUS_TYPE, decodeInt[FocusType])
}

// SetFocusType switches between the webcam's autofocus modes and manual focus.
func (c *Client) SetFocusType(ctx context.Context, t FocusType) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_FOCUS_TYPE, intValue(t))
}

// ManualFocus reads the value set by [Client.SetManualFocus]. Outside [FocusManual] it reports something else, most
// likely the lens position the autofocus picked.
func (c *Client) ManualFocus(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_MANUAL_FOCUS, decodeInt[int])
}

// SetManualFocus sets the focus distance. [ManualFocusProperty] has the range Compose offers.
//
// The webcam only uses it while the focus type is [FocusManual], and only keeps it if set in that mode. A value set
// during autofocus is gone once the focus type is back to [FocusManual].
func (c *Client) SetManualFocus(ctx context.Context, focus int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_MANUAL_FOCUS, intValue(focus))
}
