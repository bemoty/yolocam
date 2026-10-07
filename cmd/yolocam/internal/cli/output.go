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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func (o *globalOptions) print(cmd *cobra.Command, v any, human func(io.Writer) error) error {
	if o.json {
		return writeJSONLine(cmd.OutOrStdout(), v)
	}
	return human(cmd.OutOrStdout())
}

func writeJSONLine(w io.Writer, v any) error {
	return json.MarshalEncode(jsontext.NewEncoder(w), v, json.Deterministic(true))
}

func warn(cmd *cobra.Command, msg string) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "yolocam: warning: %s\n", msg)
}
