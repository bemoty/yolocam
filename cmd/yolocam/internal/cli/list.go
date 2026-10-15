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
	"maps"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
)

func newListCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all properties supported by get/set",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			listings := make([]yolocam.Property, 0, len(allProperties))
			for _, name := range propertyNames() {
				p, _ := lookupProperty(name)
				listings = append(listings, p.Property)
			}
			return opts.print(cmd, listings, func(w io.Writer) error { return writeListings(w, listings) })
		},
	}
}

func writeListings(out io.Writer, listings []yolocam.Property) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for _, l := range listings {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", l.Name, l.Summary, formatSchema(l.Value)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func formatSchema(s yolocam.Schema) string {
	switch {
	case len(s.Enum) > 0:
		values := make([]string, len(s.Enum))
		for i, v := range s.Enum {
			values[i] = fmt.Sprint(v)
		}
		return strings.Join(values, "|")
	case s.Minimum != nil && s.Maximum != nil:
		return fmt.Sprintf("%g to %g", *s.Minimum, *s.Maximum)
	case len(s.Properties) > 0:
		return "{" + strings.Join(slices.Sorted(maps.Keys(s.Properties)), ", ") + "}"
	default:
		return s.Type
	}
}
