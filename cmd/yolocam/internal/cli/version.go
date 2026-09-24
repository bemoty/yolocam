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
