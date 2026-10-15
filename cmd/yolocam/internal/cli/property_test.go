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
	"testing"

	"github.com/bemoty/yolocam"
)

func applySet(p property, raw string) error {
	action, err := p.parseSet(raw)
	if err != nil {
		return err
	}
	return action(context.Background(), nil)
}

func TestAccessor(t *testing.T) {
	var stored yolocam.ExposureType
	p := accessor(yolocam.Property{Name: "mode"},
		func(*yolocam.Client, context.Context) (yolocam.ExposureType, error) { return stored, nil },
		func(_ *yolocam.Client, _ context.Context, v yolocam.ExposureType) error { stored = v; return nil },
		parseText[yolocam.ExposureType])

	if err := applySet(p, "manual"); err != nil || stored != yolocam.ExposureManual {
		t.Fatalf("set manual = %v, stored %v", err, stored)
	}
	if got, err := p.get(context.Background(), nil); err != nil || got != yolocam.ExposureManual {
		t.Errorf("get = (%v, %v), want manual", got, err)
	}

	_, err := p.parseSet("bogus")
	if err == nil || err.Error() != `mode: invalid exposure_type "bogus", want one of: auto, manual` {
		t.Errorf("bad value err = %v", err)
	}
}

func TestEveryPropertyHasAnAccessor(t *testing.T) {
	for _, desc := range yolocam.Properties() {
		if _, ok := lookupProperty(desc.Name); !ok {
			t.Errorf("%s has no accessor", desc.Name)
		}
	}
}

func TestParseJSON(t *testing.T) {
	zoom, err := parseJSON(yolocam.CenteredZoom(0))(`{"factor":2}`)
	if err != nil || zoom != yolocam.CenteredZoom(2) {
		t.Errorf("zoom without center = (%v, %v), want centered", zoom, err)
	}
	if _, err := parseJSON[yolocam.HSL](nil)(`{"red":{"hu":10}}`); err == nil {
		t.Error("misspelled hsl channel was accepted")
	}
}
