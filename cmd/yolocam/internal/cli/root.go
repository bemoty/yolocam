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
	"time"

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
	"github.com/bemoty/yolocam/internal/firmware"
)

type globalOptions struct {
	host    string
	timeout time.Duration
	json    bool
}

func (o *globalOptions) connect(ctx context.Context) (*yolocam.Client, error) {
	return yolocam.Connect(ctx, &yolocam.Options{Host: o.host, RequestTimeout: o.timeout})
}

func (o *globalOptions) withCamera(cmd *cobra.Command, fn func(ctx context.Context, c *yolocam.Client) error) error {
	client, err := o.connect(cmd.Context())
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	return fn(cmd.Context(), client)
}

func newRootCommand() *cobra.Command {
	opts := &globalOptions{}
	root := &cobra.Command{
		Use:           "yolocam",
		Short:         "Control a YoloLiv YoloCam S3 from the command line",
		Version:       build.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := root.PersistentFlags()
	flags.StringVar(&opts.host, "host", envOr("YOLOCAM_HOST", firmware.Host), "camera address")
	flags.DurationVar(&opts.timeout, "timeout", 5*time.Second, "timeout for connecting and for each request")
	flags.BoolVar(&opts.json, "json", false, "print output as JSON")

	root.AddCommand(newGetCommand(opts))
	root.AddCommand(newSetCommand(opts))
	root.AddCommand(newWatchCommand(opts))
	root.AddCommand(newInfoCommand(opts))
	root.AddCommand(newListCommand(opts))
	root.AddCommand(newVersionCommand(opts))

	return root
}

func Execute(ctx context.Context) error {
	root := newRootCommand()
	err := root.ExecuteContext(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(root.ErrOrStderr(), "yolocam: %s\n", err)
	}
	return err
}
