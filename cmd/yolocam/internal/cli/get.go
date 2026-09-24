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
