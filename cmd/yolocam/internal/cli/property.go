package cli

import (
	"context"
	"encoding"
	"fmt"
	"slices"
	"strconv"

	"github.com/bemoty/yolocam"
)

type cameraAction func(ctx context.Context, c *yolocam.Client) error

type property struct {
	name     string
	summary  string
	get      func(ctx context.Context, c *yolocam.Client) (any, error)
	parseSet func(raw string) (cameraAction, error)
	warn     warningCheck
}

type warningCheck func(ctx context.Context, c *yolocam.Client) string

func accessor[T any](
	name, summary string,
	get func(*yolocam.Client, context.Context) (T, error),
	set func(*yolocam.Client, context.Context, T) error,
	parse func(string) (T, error),
) property {
	return property{
		name:    name,
		summary: summary,
		get: func(ctx context.Context, c *yolocam.Client) (any, error) {
			return get(c, ctx)
		},
		parseSet: func(raw string) (cameraAction, error) {
			v, err := parse(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			return func(ctx context.Context, c *yolocam.Client) error { return set(c, ctx, v) }, nil
		},
	}
}

func (p property) withWarning(warn warningCheck) property {
	p.warn = warn
	return p
}

func parseInt(raw string) (int, error) {
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%q is not an integer", raw)
	}
	return v, nil
}

func parseBool(raw string) (bool, error) {
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%q is not a boolean", raw)
	}
	return v, nil
}

func parseText[T any, PT interface {
	*T
	encoding.TextUnmarshaler
}](raw string) (T, error) {
	var v T
	err := PT(&v).UnmarshalText([]byte(raw))
	return v, err
}

var allProperties = slices.Concat(exposureProperties, whiteBalanceProperties)

func lookupProperty(name string) (property, bool) {
	for _, p := range allProperties {
		if p.name == name {
			return p, true
		}
	}
	return property{}, false
}

func propertyNames() []string {
	names := make([]string, 0, len(allProperties))
	for _, p := range allProperties {
		names = append(names, p.name)
	}
	slices.Sort(names)
	return names
}
