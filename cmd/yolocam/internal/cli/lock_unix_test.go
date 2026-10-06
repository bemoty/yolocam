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

//go:build unix

package cli

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestSessionLockIsExclusiveUntilUnlocked(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	unlock, err := lockSession()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lockSession(); !errors.Is(err, errCameraBusy) {
		t.Errorf("second lock: err = %v, want errCameraBusy", err)
	}

	unlock()
	unlock, err = lockSession()
	if err != nil {
		t.Fatalf("lock after unlock: %v", err)
	}
	unlock()
}

func TestSessionLockCreatesMissingFallbackDir(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("XDG_CACHE_HOME", filepath.Join(t.TempDir(), "not-there-yet"))

	unlock, err := lockSession()
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}
