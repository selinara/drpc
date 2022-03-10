package drpc

import (
	"time"
)

type Config struct {
	RpcTimeout       time.Duration
	HandshakeTimeout time.Duration
	PingTimeout      time.Duration
}
