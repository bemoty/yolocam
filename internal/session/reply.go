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
