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
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
)

type messageLine struct {
	Type         yolocam.MessageType `json:"type"`
	PropertyID   int                 `json:"property_id"`
	Property     string              `json:"property,omitempty"`
	Seq          uint32              `json:"seq"`
	Status       uint32              `json:"status"`
	DroppedSoFar uint64              `json:"dropped_so_far"`
	Value        json.RawMessage     `json:"value,omitempty"`
	Raw          string              `json:"raw,omitempty"`
	Error        string              `json:"error,omitempty"`
}

func newMessageLine(m yolocam.Message, alwaysRaw bool) messageLine {
	line := messageLine{
		Type:         m.Type,
		PropertyID:   m.PropertyID,
		Property:     m.PropertyName(),
		Seq:          m.Seq,
		Status:       m.Status,
		DroppedSoFar: m.DroppedSoFar,
	}
	if m.Err != nil {
		line.Error = m.Err.Error()
	} else if value, err := m.ValueJSON(); err != nil {
		line.Error = err.Error()
	} else {
		line.Value = value
	}
	if alwaysRaw || line.Error != "" {
		line.Raw = hex.EncodeToString(m.Raw)
	}
	return line
}

func newWatchCommand(opts *globalOptions) *cobra.Command {
	var alwaysRaw bool
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Stream frames the camera sends unprompted as JSON lines until interrupted",
		Long: "Stream frames the camera sends unprompted as JSON lines until interrupted: pushes, plus\n" +
			"replies that arrive after their request gave up.\n\n" +
			"Each line carries the decoded value. The raw protobuf payload (hex) is added for\n" +
			"malformed frames, or always with --raw.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.connect(cmd.Context())
			if err != nil {
				return err
			}
			defer func() { _ = client.Close() }()

			enc := json.NewEncoder(cmd.OutOrStdout())
			err = client.Watch(cmd.Context(), func(m yolocam.Message) error {
				return enc.Encode(newMessageLine(m, alwaysRaw))
			})
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&alwaysRaw, "raw", false, "include the raw protobuf payload on every line")
	return cmd
}
