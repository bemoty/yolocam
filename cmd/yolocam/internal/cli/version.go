package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Set via -X ldflags by GoReleaser; left at their zero values for `go build`/`go run`.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type buildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func (b buildInfo) String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", b.Version, b.Commit, b.Date)
}

var build = buildInfo{Version: version, Commit: commit, Date: date}

func newVersionCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the client version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.print(cmd, build, func(w io.Writer) error {
				_, err := fmt.Fprintln(w, build)
				return err
			})
		},
	}
}
