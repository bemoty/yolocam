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
