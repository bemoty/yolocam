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

var colorProperties = []property{
	accessor(yolocam.SaturationProperty,
		(*yolocam.Client).Saturation, (*yolocam.Client).SetSaturation, parseInt),
	accessor(yolocam.HSLEnabledProperty,
		(*yolocam.Client).HSLEnabled, (*yolocam.Client).SetHSLEnabled, parseBool),
	accessor(yolocam.HSLProperty,
		(*yolocam.Client).HSL, (*yolocam.Client).SetHSL, parseHSL),
}

func parseHSL(raw string) (yolocam.HSL, error) {
	h := yolocam.HSL{}
	if raw == "none" {
		return h, nil
	}

	invalid := fmt.Errorf("%q is not COLOR.CHANNEL=N,... (channel hue, saturation or lightness)", raw)
	for part := range strings.SplitSeq(raw, ",") {
		key, valueText, ok := strings.Cut(part, "=")
		colorText, channel, okKey := strings.Cut(key, ".")
		n, err := strconv.Atoi(valueText)
		if !ok || !okKey || err != nil {
			return nil, invalid
		}
		var color yolocam.HSLColor
		if err := color.UnmarshalText([]byte(colorText)); err != nil {
			return nil, err
		}

		shift := h[color]
		switch channel {
		case "hue":
			shift.Hue = n
		case "saturation":
			shift.Saturation = n
		case "lightness":
			shift.Lightness = n
		default:
			return nil, invalid
		}
		h[color] = shift
	}
	return h, nil
}
