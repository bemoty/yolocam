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
	"fmt"
	"strconv"
	"strings"

	"github.com/bemoty/yolocam"
)

var whiteBalanceProperties = []property{
	accessor(yolocam.WhiteBalanceTypeProperty,
		(*yolocam.Client).WhiteBalanceType, (*yolocam.Client).SetWhiteBalanceType, parseText),
	accessor(yolocam.WhiteBalanceKelvinProperty,
		(*yolocam.Client).WhiteBalanceKelvin, (*yolocam.Client).SetWhiteBalanceKelvin, parseInt),
	accessor(yolocam.WhiteBalanceToneProperty,
		(*yolocam.Client).WhiteBalanceTone, (*yolocam.Client).SetWhiteBalanceTone, parseInt),
	accessor(yolocam.WhiteBalanceOffsetProperty,
		(*yolocam.Client).WhiteBalanceOffset, (*yolocam.Client).SetWhiteBalanceOffset, parseWhiteBalanceOffset),
	accessor(yolocam.WhiteBalanceAutoLockProperty,
		(*yolocam.Client).WhiteBalanceAutoLock, (*yolocam.Client).SetWhiteBalanceAutoLock, parseBool),
}

func parseWhiteBalanceOffset(raw string) (yolocam.WhiteBalanceOffset, error) {
	var o yolocam.WhiteBalanceOffset
	var haveRed, haveBlue bool
	invalid := fmt.Errorf("%q is not red=N,blue=N", raw)
	for part := range strings.SplitSeq(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
		if !ok || err != nil {
			return o, invalid
		}
		switch key {
		case "red":
			o.Red, haveRed = n, true
		case "blue":
			o.Blue, haveBlue = n, true
		default:
			return o, invalid
		}
	}
	if !haveRed || !haveBlue {
		return o, invalid
	}
	return o, nil
}
