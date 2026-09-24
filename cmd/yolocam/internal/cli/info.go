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
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
	"github.com/bemoty/yolocam/internal/firmware"
)

func newInfoCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show camera model, firmware, serial and Bluetooth version",
		Long: "Show camera model, firmware, serial and Bluetooth version.\n\n" +
			"Warns on stderr if the firmware release is not one this client was verified against.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.withCamera(cmd, func(ctx context.Context, c *yolocam.Client) error {
				info, err := c.Info(ctx)
				if err != nil {
					return err
				}
				if !info.FirmwareVerified() {
					warn(cmd, fmt.Sprintf("firmware %s not verified with this client (verified: %s)",
						formatRelease(firmware.Release{Version: info.FirmwareVersion, Build: info.Build}),
						formatReleases(firmware.Verified())))
				}
				return opts.print(cmd, info, func(w io.Writer) error { return writeInfo(w, info) })
			})
		},
	}
}

func writeInfo(out io.Writer, info yolocam.DeviceInfo) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"model", info.Model},
		{"firmware", info.FirmwareVersion},
		{"build", info.Build},
		{"serial", info.Serial},
		{"bluetooth", info.BluetoothVersion},
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", row[0], row[1]); err != nil {
			return err
		}
	}
	return w.Flush()
}

func formatRelease(r firmware.Release) string {
	return fmt.Sprintf("%s build %s", r.Version, r.Build)
}

func formatReleases(rs []firmware.Release) string {
	parts := make([]string, len(rs))
	for i, r := range rs {
		parts[i] = formatRelease(r)
	}
	return strings.Join(parts, ", ")
}
