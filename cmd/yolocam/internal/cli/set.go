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
