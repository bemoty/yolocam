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
	"context"
	"encoding"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/bemoty/yolocam"
)

type cameraAction func(ctx context.Context, c *yolocam.Client) error

type property struct {
	yolocam.Property
	get      func(ctx context.Context, c *yolocam.Client) (any, error)
	parseSet func(raw string) (cameraAction, error)
}

var allProperties = slices.Concat(
	exposureProperties,
	whiteBalanceProperties,
	focusProperties,
	zoomProperties,
	imageProperties,
	colorProperties,
	deviceProperties,
)

func accessor[T any](
	desc yolocam.Property,
	get func(*yolocam.Client, context.Context) (T, error),
	set func(*yolocam.Client, context.Context, T) error,
	parse func(string) (T, error),
) property {
	return property{
		Property: desc,
		get: func(ctx context.Context, c *yolocam.Client) (any, error) {
			return get(c, ctx)
		},
		parseSet: func(raw string) (cameraAction, error) {
			v, err := parse(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", desc.Name, err)
			}
			return func(ctx context.Context, c *yolocam.Client) error { return set(c, ctx, v) }, nil
		},
	}
}

func ineffectiveWarning(ctx context.Context, c *yolocam.Client, when yolocam.Condition) string {
	if when == (yolocam.Condition{}) {
		return ""
	}
	mode, ok := lookupProperty(when.Property)
	if !ok {
		return ""
	}
	value, err := mode.get(ctx, c)
	if err != nil || fmt.Sprint(value) == when.Equals {
		return ""
	}
	return fmt.Sprintf("no effect while %s is %v", when.Property, value)
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

func parseJSON[T any](defaults T) func(string) (T, error) {
	return func(raw string) (T, error) {
		v := defaults
		if err := json.Unmarshal([]byte(raw), &v, json.RejectUnknownMembers(true)); err != nil {
			return defaults, err
		}
		return v, nil
	}
}

func (p property) parseJSONValue(raw jsontext.Value) (cameraAction, error) {
	text, err := jsonValueText(p.Value.Type, raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name, err)
	}
	return p.parseSet(text)
}

func jsonValueText(schemaType string, raw jsontext.Value) (string, error) {
	switch {
	case schemaType == "integer" && raw.Kind() == '0':
		return integerText(raw)
	case schemaType == "boolean" && (raw.Kind() == 't' || raw.Kind() == 'f'),
		schemaType == "object" && raw.Kind() == '{':
		return string(raw), nil
	case schemaType == "string" && raw.Kind() == '"':
		var text string
		err := json.Unmarshal(raw, &text)
		return text, err
	}
	return "", fmt.Errorf("%s is not a JSON %s", raw, schemaType)
}

func integerText(raw jsontext.Value) (string, error) {
	if !bytes.ContainsAny(raw, ".eE") {
		return string(raw), nil
	}
	var n float64
	if err := json.Unmarshal(raw, &n); err != nil || n != math.Trunc(n) || n < math.MinInt64 || n >= math.MaxInt64 {
		return "", fmt.Errorf("%s is not an integer", raw)
	}
	return strconv.FormatInt(int64(n), 10), nil
}

func lookupProperty(name string) (property, bool) {
	for _, p := range allProperties {
		if p.Name == name {
			return p, true
		}
	}
	return property{}, false
}

func propertyNames() []string {
	names := make([]string, 0, len(allProperties))
	for _, p := range allProperties {
		names = append(names, p.Name)
	}
	slices.Sort(names)
	return names
}
