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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bemoty/yolocam"
)

var whiteBalanceProperties = []property{
	accessor("wb-type", "auto or manual white balance",
		(*yolocam.Client).WhiteBalanceType, (*yolocam.Client).SetWhiteBalanceType, parseText),
	accessor("wb-kelvin", "white balance color temperature in kelvin, 2000 to 10000",
		(*yolocam.Client).WhiteBalanceKelvin, (*yolocam.Client).SetWhiteBalanceKelvin, parseInt,
	).withWarning(warnUnlessWhiteBalance(yolocam.WhiteBalanceManual)),
	accessor("wb-tone", "white balance tone, 0 to 128",
		(*yolocam.Client).WhiteBalanceTone, (*yolocam.Client).SetWhiteBalanceTone, parseInt,
	).withWarning(warnUnlessWhiteBalance(yolocam.WhiteBalanceManual)),
	accessor("wb-offset", "white balance red and blue offset as red=N,blue=N, each 0 to 20",
		(*yolocam.Client).WhiteBalanceOffset, (*yolocam.Client).SetWhiteBalanceOffset, parseWhiteBalanceOffset,
	).withWarning(warnUnlessWhiteBalance(yolocam.WhiteBalanceManual)),
	accessor("wb-auto-lock", "lock auto white balance at its current value, true or false",
		(*yolocam.Client).WhiteBalanceAutoLock, (*yolocam.Client).SetWhiteBalanceAutoLock, parseBool,
	).withWarning(warnUnlessWhiteBalance(yolocam.WhiteBalanceAuto)),
}

func warnUnlessWhiteBalance(want yolocam.WhiteBalanceType) warningCheck {
	return func(ctx context.Context, c *yolocam.Client) string {
		mode, err := c.WhiteBalanceType(ctx)
		if err != nil || mode == want {
			return ""
		}
		return fmt.Sprintf("no visible change while wb_type is %s", mode)
	}
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
