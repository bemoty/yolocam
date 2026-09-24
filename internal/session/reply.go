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

package session

import (
	"fmt"

	yolocamv1 "github.com/bemoty/yolocam/internal/genproto/yolocam/v1"
)

type replyKey struct {
	seq      uint32
	property yolocamv1.PropertyId
}

type MalformedError struct {
	Payload []byte
	Err     error
}

func (e *MalformedError) Error() string {
	return fmt.Sprintf("malformed message from camera (%d bytes): %v", len(e.Payload), e.Err)
}

func (e *MalformedError) Unwrap() error {
	return e.Err
}
