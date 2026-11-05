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

	"github.com/spf13/cobra"

	"github.com/bemoty/yolocam"
)

func completeProperty(withValue bool) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
		positional, _ := splitSetArgs(cmd.Flags(), args)
		switch {
		case len(positional) == 0:
			return propertyCompletions(), cobra.ShellCompDirectiveNoFileComp
		case withValue && len(positional) == 1:
			p, ok := lookupProperty(positional[0])
			if !ok {
				break
			}
			return valueCompletions(p.Value), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func propertyCompletions() []cobra.Completion {
	comps := make([]cobra.Completion, 0, len(allProperties))
	for _, name := range propertyNames() {
		p, _ := lookupProperty(name)
		comps = append(comps, cobra.CompletionWithDesc(p.Name, p.Summary))
	}
	return comps
}

func valueCompletions(s yolocam.Schema) []cobra.Completion {
	switch {
	case len(s.Enum) > 0:
		comps := make([]cobra.Completion, len(s.Enum))
		for i, v := range s.Enum {
			comps[i] = fmt.Sprint(v)
		}
		return comps
	case s.Type == "boolean":
		return []cobra.Completion{"true", "false"}
	default:
		return cobra.AppendActiveHelp(nil, formatSchema(s))
	}
}
