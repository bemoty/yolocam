package firmware

import "time"

const HeartbeatInterval = 30 * time.Second

const MaxPayload = 1 << 20
