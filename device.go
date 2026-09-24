package yolocam

import (
	"context"
	"strings"

	"github.com/bemoty/yolocam/internal/firmware"
	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

type DeviceInfo struct {
	Model            string `json:"model"`
	FirmwareVersion  string `json:"firmware_version"`
	Build            string `json:"build"`
	Serial           string `json:"serial"`
	BluetoothVersion string `json:"bluetooth_version"`
}

func (d DeviceInfo) FirmwareVerified() bool {
	return firmware.IsVerified(firmware.Release{Version: d.FirmwareVersion, Build: d.Build})
}

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
