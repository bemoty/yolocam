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

var zoomProperties = []property{
	accessor("zoom", "zoom factor 1 to 4, optionally centered at @X,Y with each 0 to 1",
		(*yolocam.Client).Zoom, (*yolocam.Client).SetZoom, parseZoom),
}

func parseZoom(raw string) (yolocam.Zoom, error) {
	invalid := fmt.Errorf("%q is not FACTOR or FACTOR@X,Y", raw)
	factorText, centerText, hasCenter := strings.Cut(raw, "@")
	factor, err := strconv.ParseFloat(factorText, 64)
	if err != nil {
		return yolocam.Zoom{}, invalid
	}
	z := yolocam.CenteredZoom(factor)
	if !hasCenter {
		return z, nil
	}

	xText, yText, ok := strings.Cut(centerText, ",")
	x, errX := strconv.ParseFloat(xText, 64)
	y, errY := strconv.ParseFloat(yText, 64)
	if !ok || errX != nil || errY != nil {
		return yolocam.Zoom{}, invalid
	}
	z.CenterX, z.CenterY = x, y
	return z, nil
}
