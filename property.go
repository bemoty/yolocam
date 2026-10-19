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

// Property describes one webcam property: its name, the values it takes, and when setting it visibly changes the
// image. Nothing in it is enforced. Setters send whatever value you give them, so a Property is there for callers
// that want to build a UI, validate input, or explain themselves.
type Property struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Value   Schema `json:"value"`
	// EffectiveWhen is the mode another property has to be in for a set to visibly change the image. Whether the
	// webcam keeps a value set outside that mode differs per property; the setter's documentation says which.
	EffectiveWhen Condition `json:"effective_when,omitzero"`
}

// Properties returns every property this library can get and set.
func Properties() []Property {
	return []Property{
		ExposureTypeProperty,
		ExposureISOProperty,
		WhiteBalanceTypeProperty,
		WhiteBalanceKelvinProperty,
		WhiteBalanceToneProperty,
		WhiteBalanceOffsetProperty,
		WhiteBalanceAutoLockProperty,
		FocusTypeProperty,
		ManualFocusProperty,
		ZoomProperty,
		SaturationProperty,
		HSLEnabledProperty,
		HSLProperty,
	}
}

// Condition is a property being set to one particular value.
type Condition struct {
	Property string `json:"property"`
	Equals   string `json:"equals"`
}

// Schema describes a value the way JSON Schema does, using only the keywords below. Ranges are the ones Compose
// offers.
type Schema struct {
	Type       string            `json:"type"`
	Minimum    *float64          `json:"minimum,omitzero"`
	Maximum    *float64          `json:"maximum,omitzero"`
	Enum       []any             `json:"enum,omitzero"`
	Properties map[string]Schema `json:"properties,omitzero"`
}
