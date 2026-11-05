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
	"bytes"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/bemoty/yolocam"
)

func newListCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all properties supported by get/set",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			listings := listedProperties()
			return opts.print(cmd, listings, func(w io.Writer) error { return writeListings(w, listings) })
		},
	}
}

func listedProperties() []yolocam.Property {
	listings := make([]yolocam.Property, 0, len(allProperties))
	for _, name := range propertyNames() {
		p, _ := lookupProperty(name)
		listings = append(listings, p.Property)
	}
	return listings
}

func writeListings(out io.Writer, listings []yolocam.Property) error {
	var table bytes.Buffer
	w := tabwriter.NewWriter(&table, 0, 4, 2, ' ', 0)
	for _, l := range listings {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", l.Name, formatListedSchema(l.Value), l.Summary)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	width := terminalWidth(out)
	for line := range strings.Lines(table.String()) {
		if _, err := io.WriteString(out, truncateLine(line, width)); err != nil {
			return err
		}
	}
	return nil
}

func terminalWidth(out io.Writer) int {
	f, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return 0
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return 0
	}
	return width
}

func truncateLine(line string, width int) string {
	const ellipsis = "..."
	runes := []rune(strings.TrimSuffix(line, "\n"))
	if width <= len(ellipsis) || len(runes) <= width {
		return line
	}
	return string(runes[:width-len(ellipsis)]) + ellipsis + "\n"
}

func formatListedSchema(s yolocam.Schema) string {
	const maxListed = 6
	switch {
	case len(s.Enum) > maxListed:
		return fmt.Sprintf("%v to %v, %d values", s.Enum[0], s.Enum[len(s.Enum)-1], len(s.Enum))
	case len(s.Properties) > maxListed:
		return fmt.Sprintf("{%d keys}", len(s.Properties))
	default:
		return formatSchema(s)
	}
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
