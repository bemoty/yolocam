package cli

import (
	"bytes"
	"testing"

	"github.com/bemoty/yolocam"
)

func TestWriteInfo(t *testing.T) {
	var out bytes.Buffer
	info := yolocam.DeviceInfo{Model: "YunxiCamera-ch131b", FirmwareVersion: "1.0.0", Build: "1295", Serial: "S", BluetoothVersion: "26030202"}
	if err := writeInfo(&out, info); err != nil {
		t.Fatal(err)
	}
	want := "model      YunxiCamera-ch131b\n" +
		"firmware   1.0.0\n" +
		"build      1295\n" +
		"serial     S\n" +
		"bluetooth  26030202\n"
	if out.String() != want {
		t.Errorf("got\n%s\nwant\n%s", out.String(), want)
	}
}
