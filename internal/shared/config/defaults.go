package config

import "time"

const (
	DefaultHost         = "127.0.0.1"
	DefaultServerHost   = "0.0.0.0"
	DefaultGRPCPort     = 50051
	DefaultHTTPPort     = 8080
	DefaultWebPort      = 9091
	DefaultPageSize     = 100
	DefaultMaxPageSize  = 10000
	DefaultSessionTTL   = 7 * 24 * time.Hour
	DefaultAccessTTL    = 15 * time.Minute
	DefaultRefreshTTL   = 7 * 24 * time.Hour
	DefaultReadTimeout  = 15 * time.Second
	DefaultWriteTimeout = 15 * time.Second
	DefaultIdleTimeout  = 60 * time.Second
	DefaultShutdownTime = 10 * time.Second
)
