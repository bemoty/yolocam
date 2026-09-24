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
	"io"

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
)

type propertyValue struct {
	Property string `json:"property"`
	Value    any    `json:"value"`
}

func newGetCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:       "get <property>",
		Short:     "Read a camera property",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: propertyNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			p, _ := lookupProperty(name)

			return opts.withCamera(cmd, func(ctx context.Context, c *yolocam.Client) error {
				value, err := p.get(ctx, c)
				if err != nil {
					return err
				}
				return opts.print(cmd, propertyValue{Property: name, Value: value}, func(w io.Writer) error {
					_, err := fmt.Fprintln(w, value)
					return err
				})
			})
		},
	}
}
