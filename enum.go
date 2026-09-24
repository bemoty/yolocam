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

package yolocam

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

type enum[T ~int] struct {
	kind  string
	names map[T]string
}

func (e enum[T]) format(v T) string {
	if name, ok := e.names[v]; ok {
		return name
	}
	return fmt.Sprintf("%s(%d)", e.kind, int(v))
}

func (e enum[T]) marshal(v T) []byte {
	if name, ok := e.names[v]; ok {
		return []byte(name)
	}
	return strconv.AppendInt(nil, int64(v), 10)
}

func (e enum[T]) unmarshal(text []byte) (T, error) {
	for v, name := range e.names {
		if name == string(text) {
			return v, nil
		}
	}
	return 0, fmt.Errorf("invalid %s %q, want one of: %s", e.kind, text, strings.Join(e.validNames(), ", "))
}

func (e enum[T]) validNames() []string {
	names := make([]string, 0, len(e.names))
	for _, v := range slices.Sorted(maps.Keys(e.names)) {
		names = append(names, e.names[v])
	}
	return names
}
