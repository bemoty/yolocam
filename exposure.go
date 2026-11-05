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

// ExposureType defines whether the webcam picks the exposure itself or uses the ISO you set.
type ExposureType int

const (
	ExposureAuto   ExposureType = 0
	ExposureManual ExposureType = 1
)

var exposureTypes = enum[ExposureType]{kind: "exposure_type", names: map[ExposureType]string{
	ExposureAuto:   "auto",
	ExposureManual: "manual",
}}

func (t ExposureType) String() string {
	return exposureTypes.format(t)
}

func (t ExposureType) MarshalText() ([]byte, error) {
	return exposureTypes.marshal(t), nil
}

func (t *ExposureType) UnmarshalText(text []byte) error {
	v, err := exposureTypes.unmarshal(text)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

var ExposureTypeProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE),
	Summary: "whether the webcam picks the exposure itself",
	Value:   exposureTypes.schema(),
}

// ExposureType reads the value set by [Client.SetExposureType].
func (c *Client) ExposureType(ctx context.Context) (ExposureType, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE, decodeInt[ExposureType])
}

// SetExposureType switches between auto and manual exposure.
func (c *Client) SetExposureType(ctx context.Context, t ExposureType) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE, intValue(t))
}

var ExposureISOProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO),
	Summary:       "ISO sensitivity in third stops",
	Value:         Schema{Type: "integer", Enum: []any{100, 125, 160, 200, 250, 320, 400, 500, 640, 800, 1000, 1250, 1600, 2000, 2500, 3200, 4000, 5000, 6400}},
	EffectiveWhen: Condition{Property: ExposureTypeProperty.Name, Equals: ExposureManual.String()},
}

// ExposureISO reads the value set by [Client.SetExposureISO].
func (c *Client) ExposureISO(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO, decodeInt[int])
}

// SetExposureISO sets the webcam's ISO sensitivity. [ExposureISOProperty] has the values Compose offers.
//
// While the exposure type is [ExposureAuto], the webcam ignores this: the set succeeds, but the stored value doesn't
// change. Switch to [ExposureManual] first. The value set in manual is kept across switches to auto and back.
func (c *Client) SetExposureISO(ctx context.Context, iso int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO, intValue(iso))
}

var ShutterProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_SHUTTER_US),
	Summary:       "shutter time in microseconds, 1/6400 s to 1/60 s rounded down",
	Value:         Schema{Type: "integer", Enum: []any{156, 200, 250, 312, 400, 500, 625, 800, 1000, 1250, 1562, 2000, 2500, 3125, 4000, 5000, 6250, 8000, 10000, 12500, 16666}},
	EffectiveWhen: Condition{Property: ExposureTypeProperty.Name, Equals: ExposureManual.String()},
}

// Shutter reads the value set by [Client.SetShutter], in microseconds.
func (c *Client) Shutter(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_SHUTTER_US, decodeInt[int])
}

// SetShutter sets the shutter time in microseconds. [ShutterProperty] has the values Compose offers.
//
// While the exposure type is [ExposureAuto], the webcam ignores this: the set succeeds, but the stored value doesn't
// change. Switch to [ExposureManual] first.
func (c *Client) SetShutter(ctx context.Context, micros int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_SHUTTER_US, intValue(micros))
}

var ExposureAutoLockProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_AUTO_LOCK),
	Summary:       "whether the auto exposure is locked at its current value",
	Value:         Schema{Type: "boolean"},
	EffectiveWhen: Condition{Property: ExposureTypeProperty.Name, Equals: ExposureAuto.String()},
}

// ExposureAutoLock reports whether the auto exposure is locked at its current value.
func (c *Client) ExposureAutoLock(ctx context.Context) (bool, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_AUTO_LOCK, decodeBool)
}

// SetExposureAutoLock locks or unlocks the auto exposure at its current value. It only does something while the
// exposure type is [ExposureAuto], but a value set in manual is kept. Compose hides anti-flicker and EV while it is
// locked.
func (c *Client) SetExposureAutoLock(ctx context.Context, locked bool) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_AUTO_LOCK, boolValue(locked))
}

// AntiFlicker is the mains frequency the auto exposure avoids flicker for.
type AntiFlicker int

const (
	AntiFlickerOff  AntiFlicker = 0
	AntiFlicker50Hz AntiFlicker = 1
	AntiFlicker60Hz AntiFlicker = 2
)

var antiFlickers = enum[AntiFlicker]{kind: "anti_flicker", names: map[AntiFlicker]string{
	AntiFlickerOff:  "off",
	AntiFlicker50Hz: "50hz",
	AntiFlicker60Hz: "60hz",
}}

func (a AntiFlicker) String() string {
	return antiFlickers.format(a)
}

func (a AntiFlicker) MarshalText() ([]byte, error) {
	return antiFlickers.marshal(a), nil
}

func (a *AntiFlicker) UnmarshalText(text []byte) error {
	v, err := antiFlickers.unmarshal(text)
	if err != nil {
		return err
	}
	*a = v
	return nil
}

var AntiFlickerProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_ANTI_FLICKER),
	Summary:       "mains frequency the auto exposure avoids flicker for",
	Value:         antiFlickers.schema(),
	EffectiveWhen: Condition{Property: ExposureTypeProperty.Name, Equals: ExposureAuto.String()},
}

// AntiFlicker reads the value set by [Client.SetAntiFlicker].
func (c *Client) AntiFlicker(ctx context.Context) (AntiFlicker, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_ANTI_FLICKER, decodeInt[AntiFlicker])
}

// SetAntiFlicker sets the mains frequency the auto exposure avoids flicker for.
//
// While the exposure type is [ExposureManual], the webcam ignores this: the set succeeds, but the stored value
// doesn't change. Switch to [ExposureAuto] first.
func (c *Client) SetAntiFlicker(ctx context.Context, a AntiFlicker) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_ANTI_FLICKER, intValue(a))
}

const (
	MinEV = -8
	MaxEV = 8
)

var EVProperty = Property{
	Name:          propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_EV),
	Summary:       "exposure compensation, higher is brighter",
	Value:         Schema{Type: "integer", Minimum: new(float64(MinEV)), Maximum: new(float64(MaxEV))},
	EffectiveWhen: Condition{Property: ExposureTypeProperty.Name, Equals: ExposureAuto.String()},
}

// EV reads the value set by [Client.SetEV].
func (c *Client) EV(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EV, decodeEV)
}

// SetEV sets the auto exposure's compensation, from [MinEV] to [MaxEV].
//
// While the exposure type is [ExposureManual], the webcam ignores this: the set succeeds, but the stored value
// doesn't change. Switch to [ExposureAuto] first.
func (c *Client) SetEV(ctx context.Context, ev int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EV, evValue(ev))
}

func evValue(ev int) *yolocamv1.Value {
	return intValue(ev - MinEV)
}

func decodeEV(v *yolocamv1.Value) (int, error) {
	n, err := decodeInt[int](v)
	if err != nil {
		return 0, err
	}
	return n + MinEV, nil
}
