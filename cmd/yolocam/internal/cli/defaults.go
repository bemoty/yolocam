package cli

import "os"

const fallbackHost = "192.168.123.10"

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
