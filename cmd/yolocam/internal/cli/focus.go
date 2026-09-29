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

var focusProperties = []property{
	accessor("focus-type", "af-s, af-c, face or mf (manual)",
		(*yolocam.Client).FocusType, (*yolocam.Client).SetFocusType, parseText),
	accessor("focus", "manual focus distance, 0 to 100",
		(*yolocam.Client).ManualFocus, (*yolocam.Client).SetManualFocus, parseInt,
	).withWarning(warnUnlessManualFocus),
}

func warnUnlessManualFocus(ctx context.Context, c *yolocam.Client) string {
	mode, err := c.FocusType(ctx)
	if err != nil || mode == yolocam.FocusManual {
		return ""
	}
	return fmt.Sprintf("no visible change while focus_type is %s", mode)
}
