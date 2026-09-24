package yolocam

import (
	"context"
	"testing"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

func deviceInfoValue(modelFirmware, serial, build string) *yolocamv1.Value {
	return &yolocamv1.Value{Kind: &yolocamv1.Value_DeviceInfo{DeviceInfo: &yolocamv1.DeviceInfo{
		ModelFirmware: modelFirmware, Serial: serial, Build: build,
	}}}
}

func TestInfo(t *testing.T) {
	c, cam := newTestClient(t)

	pending := async(func() (DeviceInfo, error) { return c.Info(context.Background()) })
	req := cam.Recv()
	if req.GetType() != yolocamv1.MessageType_MESSAGE_TYPE_GET || req.GetProperty() != yolocamv1.PropertyId_PROPERTY_ID_DEVICE_INFO {
		t.Fatalf("request = %v, want GET device_info", req)
	}
	cam.Reply(req, statusOK, deviceInfoValue("YunxiCamera-ch131b-1.0.0", "SERIAL", "1295"))
	req = cam.Recv()
	if req.GetType() != yolocamv1.MessageType_MESSAGE_TYPE_GET || req.GetProperty() != yolocamv1.PropertyId_PROPERTY_ID_BLUETOOTH_VERSION {
		t.Fatalf("request = %v, want GET bluetooth_version", req)
	}
	cam.Reply(req, statusOK, &yolocamv1.Value{Kind: &yolocamv1.Value_StringValue_2{StringValue_2: "26030202"}})

	r := await(t, pending)
	want := DeviceInfo{Model: "YunxiCamera-ch131b", FirmwareVersion: "1.0.0", Build: "1295", Serial: "SERIAL", BluetoothVersion: "26030202"}
	if r.err != nil || r.v != want {
		t.Fatalf("Info = (%+v, %v), want %+v", r.v, r.err, want)
	}
	if !r.v.FirmwareVerified() {
		t.Error("1.0.0 build 1295 not reported as verified")
	}
}

func TestDecodeDeviceInfo(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		build        string
		wantModel    string
		wantVersion  string
		wantVerified bool
	}{
		{"verified", "YunxiCamera-ch131b-1.0.0", "1295", "YunxiCamera-ch131b", "1.0.0", true},
		{"other build", "YunxiCamera-ch131b-1.0.0", "1300", "YunxiCamera-ch131b", "1.0.0", false},
		{"other version", "YunxiCamera-ch131b-1.1.0", "1295", "YunxiCamera-ch131b", "1.1.0", false},
		{"no separator", "YunxiCamera", "1295", "YunxiCamera", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := decodeDeviceInfo(deviceInfoValue(tt.raw, "", tt.build))
			if err != nil {
				t.Fatal(err)
			}
			if info.Model != tt.wantModel || info.FirmwareVersion != tt.wantVersion {
				t.Errorf("got (%q, %q), want (%q, %q)", info.Model, info.FirmwareVersion, tt.wantModel, tt.wantVersion)
			}
			if got := info.FirmwareVerified(); got != tt.wantVerified {
				t.Errorf("FirmwareVerified = %v, want %v", got, tt.wantVerified)
			}
		})
	}
}
