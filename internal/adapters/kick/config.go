package kick

import (
	"github.com/llan0/multichat/internal/adapters/common"
)

type Config struct {
	Channel      string
	HTTPClient   common.HTTPClient
	RetryConfig  common.RetryConfig
	WebSocketURL string
	APIURL       string
}

func DefaultConfig(channel string) Config {
	return Config{
		Channel:      channel,
		HTTPClient:   common.DefaultHTTPClient(),
		RetryConfig:  common.DefaultRetryConfig(),
		WebSocketURL: "wss://ws-us2.pusher.com/app/32cbd69e4b950bf97679?protocol=7&client=js&version=8.4.0-rc2&flash=false",
		APIURL:       "https://kick.com/api/v1/channels/",
	}
}

func (c Config) WithHTTPClient(client common.HTTPClient) Config {
	c.HTTPClient = client
	return c
}

func (c Config) WithRetryConfig(cfg common.RetryConfig) Config {
	c.RetryConfig = cfg
	return c
}
