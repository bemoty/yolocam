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
	"strings"

	"google.golang.org/protobuf/proto"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

func decodeInt[T ~int](v *yolocamv1.Value) (T, error) {
	kind, ok := v.GetKind().(*yolocamv1.Value_IntValue)
	if !ok {
		return 0, unsupportedValue(v)
	}
	return T(kind.IntValue), nil
}

func intValue[T ~int](v T) *yolocamv1.Value {
	return &yolocamv1.Value{Kind: &yolocamv1.Value_IntValue{IntValue: int64(v)}}
}

func decodeBool(v *yolocamv1.Value) (bool, error) {
	n, err := decodeInt[int](v)
	if err != nil {
		return false, err
	}
	switch n {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("%w: int %d is not a boolean", ErrUnsupportedValue, n)
	}
}

func boolValue(b bool) *yolocamv1.Value {
	if b {
		return intValue(1)
	}
	return intValue(0)
}

func unsupportedValue(v *yolocamv1.Value) error {
	raw, _ := proto.Marshal(v)
	return &UnsupportedValueError{Kind: valueKindName(v), Raw: raw}
}

func valueKindName(v *yolocamv1.Value) string {
	m := v.ProtoReflect()
	if !m.IsValid() {
		return "none"
	}
	if field := m.WhichOneof(m.Descriptor().Oneofs().ByName("kind")); field != nil {
		return string(field.Name())
	}
	if len(m.GetUnknown()) > 0 {
		return "unknown"
	}
	return "none"
}

func propertyName(id yolocamv1.PropertyId) (string, bool) {
	name, ok := yolocamv1.PropertyId_name[int32(id)]
	if !ok {
		return "", false
	}
	return strings.ToLower(strings.TrimPrefix(name, "PROPERTY_ID_")), true
}

func propertyLabel(id yolocamv1.PropertyId) string {
	if name, ok := propertyName(id); ok {
		return name
	}
	return fmt.Sprintf("property(%d)", int32(id))
}

func decodeString2(v *yolocamv1.Value) (string, error) {
	kind, ok := v.GetKind().(*yolocamv1.Value_StringValue_2)
	if !ok {
		return "", unsupportedValue(v)
	}
	return kind.StringValue_2, nil
}
