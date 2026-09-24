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

// ExposureType reads the value set by [Client.SetExposureType].
func (c *Client) ExposureType(ctx context.Context) (ExposureType, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE, decodeInt[ExposureType])
}

// SetExposureType switches between auto and manual exposure.
func (c *Client) SetExposureType(ctx context.Context, t ExposureType) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_TYPE, intValue(t))
}

// ExposureISO reads the value set by [Client.SetExposureISO].
func (c *Client) ExposureISO(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO, decodeInt[int])
}

// SetExposureISO sets the webcam's ISO sensitivity. Compose offers 100 to 6400 in third stops (100, 125, 160, 200,
// ...).
//
// Setting this value has no visible effect while the exposure type is [ExposureAuto]. The webcam still remembers the
// value set here though, so setting it to [ExposureManual] later on will read the value set here.
func (c *Client) SetExposureISO(ctx context.Context, iso int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_EXPOSURE_ISO, intValue(iso))
}
