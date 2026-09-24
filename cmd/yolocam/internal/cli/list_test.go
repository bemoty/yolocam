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
