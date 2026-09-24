package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), err
}

func TestSetRejectsBadInputBeforeConnecting(t *testing.T) {
	const unroutable = "--host=192.0.2.1"
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"set unknown property", []string{"set", "bogus", "1", unroutable}, `set: unknown property "bogus"`},
		{"set bad integer", []string{"set", "exposure", "4x", unroutable}, `exposure: "4x" is not an integer`},
		{"set bad enum", []string{"set", "exposure-type", "bogus", unroutable}, `exposure-type: invalid exposure_type "bogus"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := run(t, tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
