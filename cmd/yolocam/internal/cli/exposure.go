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

	"github.com/bemoty/yolocam"
)

var exposureProperties = []property{
	accessor("exposure", "ISO sensitivity, 100 to 6400 in third stops",
		(*yolocam.Client).ExposureISO, (*yolocam.Client).SetExposureISO, parseInt,
	).withWarning(warnIfAutoExposure),
	accessor("exposure-type", "auto or manual exposure",
		(*yolocam.Client).ExposureType, (*yolocam.Client).SetExposureType, parseText),
}

func warnIfAutoExposure(ctx context.Context, c *yolocam.Client) string {
	mode, err := c.ExposureType(ctx)
	if err != nil || mode != yolocam.ExposureAuto {
		return ""
	}
	return fmt.Sprintf("no visible change while exposure_type is %s", mode)
}
