package config

import (
	"github.com/BurntSushi/toml"
)

type IPRateLimiterSection struct {
	Enable      bool `toml:"enable"`
	MaxRequests int  `toml:"max_requests"`
	WindowMS    int  `toml:"window_ms"`
}

type IPRateLimiterConfig struct {
	Register IPRateLimiterSection `toml:"register"`
	Login    IPRateLimiterSection `toml:"login"`
	Logout   IPRateLimiterSection `toml:"logout"`
	Profile  IPRateLimiterSection `toml:"profile"`
}

type AccountPerIPLimiterConfig struct {
	Enable      bool `toml:"enable"`
	MaxAccounts int  `toml:"max_accounts"`
}

type Config struct {
	AccountPerIPLimiter AccountPerIPLimiterConfig `toml:"account_per_ip_limiter"`
	IPRateLimiter       IPRateLimiterConfig       `toml:"ip_rate_limiter"`
	DBPath              string                    `toml:"dbpath"`
	Debug               bool                      `toml:"debug"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
