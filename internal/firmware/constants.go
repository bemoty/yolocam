package firmware

import "time"

const (
	Host = "192.168.123.10"
	Port = 12345
)

const HeartbeatInterval = 30 * time.Second

const MaxPayload = 1 << 20
