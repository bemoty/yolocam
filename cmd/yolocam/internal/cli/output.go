package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func (o *globalOptions) print(cmd *cobra.Command, v any, human func(io.Writer) error) error {
	if o.json {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(v)
	}
	return human(cmd.OutOrStdout())
}

func warn(cmd *cobra.Command, msg string) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "yolocam: warning: %s\n", msg)
}
