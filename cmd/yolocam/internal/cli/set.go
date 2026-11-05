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
	"strings"

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
			positional, _ := splitSetArgs(cmd.Flags(), args)
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

func splitSetArgs(flags *pflag.FlagSet, args []string) (positional, flagArgs []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch flag, isFlag := lookupFlag(flags, arg); {
		case arg == "--" && i+1 < len(args):
			i++
			positional = append(positional, args[i])
		case isFlag:
			flagArgs = append(flagArgs, arg)
			if flag != nil && flag.NoOptDefVal == "" && !strings.Contains(arg, "=") && i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		default:
			positional = append(positional, arg)
		}
	}
	return positional, flagArgs
}

func lookupFlag(flags *pflag.FlagSet, arg string) (*pflag.Flag, bool) {
	if name, ok := strings.CutPrefix(arg, "--"); ok && name != "" {
		name, _, _ = strings.Cut(name, "=")
		return flags.Lookup(name), true
	}
	if len(arg) == 2 && arg[0] == '-' {
		flag := flags.ShorthandLookup(arg[1:])
		return flag, flag != nil
	}
	return nil, false
}

func validateSetArgs(cmd *cobra.Command, args []string) error {
	positional, flagArgs := splitSetArgs(cmd.Flags(), args)
	if err := cmd.Flags().Parse(flagArgs); err != nil {
		return err
	}
	if help, _ := cmd.Flags().GetBool("help"); help {
		return pflag.ErrHelp
	}
	return cobra.ExactArgs(2)(cmd, slices.Concat(positional, cmd.Flags().Args()))
}
