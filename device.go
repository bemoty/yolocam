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
