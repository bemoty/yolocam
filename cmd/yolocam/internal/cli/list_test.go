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
	"encoding/json"
	"testing"
)

func TestListJSON(t *testing.T) {
	out, err := run(t, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var listings []propertyListing
	if err := json.Unmarshal([]byte(out), &listings); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if len(listings) != len(allProperties) {
		t.Errorf("got %d listings, want %d", len(listings), len(allProperties))
	}
}
