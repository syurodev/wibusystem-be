package config

import (
	"errors"
	"fmt"
	"time"
)

// ServerConfig captures tunables for spinning up a gRPC server that exposes
// the token validation API. Values are typically loaded from environment
// variables via the service-specific config package.
type ServerConfig struct {
	Host                  string        `json:"host"`
	Port                  int           `json:"port"`
	MaxRecvMsgSize        int           `json:"max_recv_msg_size"`
	MaxSendMsgSize        int           `json:"max_send_msg_size"`
	ConnectionTimeout     time.Duration `json:"connection_timeout"`
	MaxConnectionIdle     time.Duration `json:"max_connection_idle"`
	MaxConnectionAge      time.Duration `json:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `json:"max_connection_age_grace"`
	Time                  time.Duration `json:"keepalive_time"`
	Timeout               time.Duration `json:"keepalive_timeout"`
	EnableReflection      bool          `json:"enable_reflection"`
}

// Address returns the listen address in host:port form. If Host is empty the
// gRPC server will listen on all interfaces.
func (c *ServerConfig) Address() string {
	host := c.Host
	if host == "" {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", host, c.Port)
}

// Validate performs basic sanity checks to catch obvious misconfiguration.
func (c *ServerConfig) Validate() error {
	if c == nil {
		return errors.New("server config is nil")
	}
	if c.Port <= 0 {
		return errors.New("grpc port must be greater than zero")
	}
	return nil
}

// WithDefaults returns a copy of the config with reasonable defaults applied
// when optional values are not provided. This avoids leaking zero values into
// grpc.NewServer options which can otherwise disable internal safeguards.
func (c *ServerConfig) WithDefaults() *ServerConfig {
	if c == nil {
		return nil
	}
	cfg := *c
	if cfg.MaxRecvMsgSize == 0 {
		cfg.MaxRecvMsgSize = 4 * 1024 * 1024
	}
	if cfg.MaxSendMsgSize == 0 {
		cfg.MaxSendMsgSize = 4 * 1024 * 1024
	}
	if cfg.ConnectionTimeout == 0 {
		cfg.ConnectionTimeout = 5 * time.Second
	}
	if cfg.MaxConnectionIdle == 0 {
		cfg.MaxConnectionIdle = 15 * time.Second
	}
	if cfg.MaxConnectionAge == 0 {
		cfg.MaxConnectionAge = 30 * time.Second
	}
	if cfg.MaxConnectionAgeGrace == 0 {
		cfg.MaxConnectionAgeGrace = 5 * time.Second
	}
	if cfg.Time == 0 {
		cfg.Time = 10 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 3 * time.Second
	}
	return &cfg
}
