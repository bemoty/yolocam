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
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bemoty/yolocam"
)

func newSetCommand(opts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "set <property> <value>",
		Short:     "Set a camera property",
		Args:      validateSetArgs,
		ValidArgs: propertyNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			positional, _ := splitSetArgs(args)
			name, value := positional[0], positional[1]
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
				if msg := ineffectiveWarning(ctx, c, p.EffectiveWhen); msg != "" {
					warn(cmd, msg)
				}
				return nil
			})
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func splitSetArgs(args []string) (positional, trailing []string) {
	if len(args) > 2 && args[1] == "--" {
		args = slices.Concat(args[:1], args[2:])
	}
	if len(args) <= 2 {
		return args, nil
	}
	return args[:2], args[2:]
}

func validateSetArgs(cmd *cobra.Command, args []string) error {
	positional, trailing := splitSetArgs(args)
	if err := cmd.Flags().Parse(trailing); err != nil {
		return err
	}
	if help, _ := cmd.Flags().GetBool("help"); help {
		return pflag.ErrHelp
	}
	return cobra.ExactArgs(2)(cmd, slices.Concat(positional, cmd.Flags().Args()))
}
