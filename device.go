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
	"strings"

	"github.com/bemoty/yolocam/internal/firmware"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

// DeviceInfo describes the webcam and the firmware it runs, as reported by [Client.Info].
type DeviceInfo struct {
	Model            string `json:"model"`            // internal name, e.g. "YunxiCamera-ch131b" for the YoloCam S3
	FirmwareVersion  string `json:"firmware_version"` // e.g. "1.0.0"; empty if it can't be split from Model
	Build            string `json:"build"`
	Serial           string `json:"serial"`
	BluetoothVersion string `json:"bluetooth_version"` // e.g. "26030202"
}

// FirmwareVerified reports whether this library was tested against exactly this firmware version and build. The
// library does not hard reject unverified firmware, but you can use this to warn users about it (and any shenanigans
// that may occur).
func (d DeviceInfo) FirmwareVerified() bool {
	return firmware.IsVerified(firmware.Release{Version: d.FirmwareVersion, Build: d.Build})
}

// Info reads the webcam's model, firmware, and serial number. Internally, that's two requests to the webcam, so it's
// a bit slower than the other getters.
func (c *Client) Info(ctx context.Context) (DeviceInfo, error) {
	info, err := c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_DEVICE_INFO, decodeDeviceInfo)
	if err != nil {
		return DeviceInfo{}, err
	}
	info.BluetoothVersion, err = c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_BLUETOOTH_VERSION, decodeString2)
	if err != nil {
		return DeviceInfo{}, err
	}
	return info, nil
}

func decodeDeviceInfo(v *yolocamv1.Value) (DeviceInfo, error) {
	kind, ok := v.GetKind().(*yolocamv1.Value_DeviceInfo)
	if !ok {
		return DeviceInfo{}, unsupportedValue(v)
	}
	raw := kind.DeviceInfo
	info := DeviceInfo{Model: raw.GetModelFirmware(), Build: raw.GetBuild(), Serial: raw.GetSerial()}
	if i := strings.LastIndexByte(info.Model, '-'); i >= 0 {
		info.Model, info.FirmwareVersion = info.Model[:i], info.Model[i+1:]
	}
	return info, nil
}

var FrameRateProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_FRAME_RATE),
	Summary: "frames per second of Compose's preview; other video apps pick their own rate",
	Value:   Schema{Type: "integer", Enum: []any{25, 30, 50, 60}},
}

// FrameRate reads the value set by [Client.SetFrameRate].
func (c *Client) FrameRate(ctx context.Context) (int, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_FRAME_RATE, decodeInt[int])
}

// SetFrameRate sets the frame rate. [FrameRateProperty] has the values Compose offers. Compose says it only affects
// its own preview, since other video apps choose a rate themselves.
func (c *Client) SetFrameRate(ctx context.Context, fps int) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_FRAME_RATE, intValue(fps))
}

var AlwaysOnProperty = Property{
	Name:    propertyLabel(yolocamv1.PropertyId_PROPERTY_ID_ALWAYS_ON),
	Summary: "whether the webcam stays on instead of turning off a minute after the last app closes it",
	Value:   Schema{Type: "boolean"},
}

// AlwaysOn reports whether the webcam stays on when no app uses it.
func (c *Client) AlwaysOn(ctx context.Context) (bool, error) {
	return c.get(ctx, yolocamv1.PropertyId_PROPERTY_ID_ALWAYS_ON, decodeBool)
}

// SetAlwaysOn keeps the webcam on, or lets it turn off a minute after the last app stops using it.
func (c *Client) SetAlwaysOn(ctx context.Context, on bool) error {
	return c.set(ctx, yolocamv1.PropertyId_PROPERTY_ID_ALWAYS_ON, boolValue(on))
}
