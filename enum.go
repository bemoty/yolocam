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
