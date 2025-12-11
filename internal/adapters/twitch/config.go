package twitch

import (
	"strings"

	"github.com/llan0/multichat/internal/adapters/common"
)

type Config struct {
	Channel     string
	RetryConfig common.RetryConfig
}

func DefaultConfig(channel string) Config {
	return Config{
		Channel:     strings.ToLower(channel),
		RetryConfig: common.DefaultRetryConfig(),
	}
}

func (c Config) WithRetryConfig(cfg common.RetryConfig) Config {
	c.RetryConfig = cfg
	return c
}
