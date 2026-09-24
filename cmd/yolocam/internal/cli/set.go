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

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
)

func newSetCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:       "set <property> <value>",
		Short:     "Set a camera property",
		Args:      cobra.ExactArgs(2),
		ValidArgs: propertyNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, value := args[0], args[1]
			p, ok := lookupProperty(name)
			if !ok {
				return fmt.Errorf("set: unknown property %q", name)
			}
			apply, err := p.parseSet(value)
			if err != nil {
				return err
			}

			return opts.withCamera(cmd, func(ctx context.Context, c *yolocam.Client) error {
				if err := apply(ctx, c); err != nil {
					return err
				}
				if p.warn == nil {
					return nil
				}
				if msg := p.warn(ctx, c); msg != "" {
					warn(cmd, msg)
				}
				return nil
			})
		},
	}
}
