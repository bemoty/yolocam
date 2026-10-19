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
	_ "embed"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"strconv"

	"github.com/spf13/cobra"
)

//go:embed openrpc.json
var openRPCTemplate []byte

func newOpenRPCCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "openrpc",
		Short: "Print the OpenRPC document describing serve --stdio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := openRPCDocument()
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(append(doc, '\n'))
			return err
		},
	}
}

func openRPCDocument() ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(openRPCTemplate, &doc); err != nil {
		return nil, err
	}
	doc["info"].(map[string]any)["version"] = strconv.Itoa(rpcProtocolVersion)

	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	var names, values []any
	for _, p := range listedProperties() {
		schemas[p.Name] = p.Value
		names = append(names, p.Name)
		values = append(values, map[string]any{"$ref": "#/components/schemas/" + p.Name})
	}
	schemas["PropertyName"].(map[string]any)["enum"] = names
	schemas["PropertyValue"].(map[string]any)["oneOf"] = values
	return json.Marshal(doc, json.Deterministic(true), jsontext.WithIndent("  "))
}
