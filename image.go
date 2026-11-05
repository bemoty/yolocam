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

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

const (
	MinSharpness = -10
	MaxSharpness = 10

	negativeSharpnessBase = 60
)

var SharpnessProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_SHARPNESS),
	Summary: "image sharpness, 0 is the default",
	Value:   Schema{Type: "integer", Minimum: new(float64(MinSharpness)), Maximum: new(float64(MaxSharpness))},
}

// Sharpness reads the value set by [Client.SetSharpness].
func (c *Client) Sharpness(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_SHARPNESS, decodeSharpness)
}

// SetSharpness sets the image sharpness, from [MinSharpness] to [MaxSharpness]. A value above 60 is rejected before
// anything is sent, as the webcam would read it back as a negative sharpness.
func (c *Client) SetSharpness(ctx context.Context, sharpness int) error {
	if sharpness > negativeSharpnessBase {
		return propertyError("set", yolocamv1.PropertyId_PROPERTY_ID_SHARPNESS,
			fmt.Errorf("%w: sharpness %d is above %d", ErrInvalidValue, sharpness, negativeSharpnessBase))
	}
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_SHARPNESS, sharpnessValue(sharpness))
}

func sharpnessValue(sharpness int) *yolocamv1.Value {
	if sharpness < 0 {
		return intValue(negativeSharpnessBase - sharpness)
	}
	return intValue(sharpness)
}

func decodeSharpness(v *yolocamv1.Value) (int, error) {
	n, err := decodeInt[int](v)
	if err != nil {
		return 0, err
	}
	if n > negativeSharpnessBase {
		return negativeSharpnessBase - n, nil
	}
	return n, nil
}

var ContrastProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_CONTRAST),
	Summary: "image contrast",
	Value:   Schema{Type: "integer", Minimum: new(0.0), Maximum: new(100.0)},
}

// Contrast reads the value set by [Client.SetContrast].
func (c *Client) Contrast(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_CONTRAST, decodeInt[int])
}

// SetContrast sets the image contrast. [ContrastProperty] has the range Compose offers.
func (c *Client) SetContrast(ctx context.Context, contrast int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_CONTRAST, intValue(contrast))
}

var DRProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_DR),
	Summary:       "what Compose labels D-R, offered only in manual exposure",
	Value:         Schema{Type: "integer", Minimum: new(0.0), Maximum: new(100.0)},
	EffectiveWhen: Condition{Property: ExposureTypeProperty.Name, Equals: ExposureManual.String()},
}

// DR reads the value set by [Client.SetDR].
func (c *Client) DR(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_DR, decodeInt[int])
}

// SetDR sets what Compose labels D-R. [DRProperty] has the range Compose offers.
//
// It only changes the image while the exposure type is [ExposureManual]. A value set in auto is kept for manual.
func (c *Client) SetDR(ctx context.Context, dr int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_DR, intValue(dr))
}

var VignettingCorrectionProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_VIGNETTING_CORRECTION),
	Summary: "whether the webcam brightens the image edges to hide the lens vignette",
	Value:   Schema{Type: "boolean"},
}

// VignettingCorrection reads the value set by [Client.SetVignettingCorrection].
func (c *Client) VignettingCorrection(ctx context.Context) (bool, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_VIGNETTING_CORRECTION, decodeBool)
}

// SetVignettingCorrection turns the lens vignette correction on or off.
func (c *Client) SetVignettingCorrection(ctx context.Context, enabled bool) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_VIGNETTING_CORRECTION, boolValue(enabled))
}
