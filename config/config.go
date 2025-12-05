package config

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aabbtree77/schatzhauser/protect"
)

type IPRateLimiterSection struct {
	Threshold int `toml:"threshold"`
	WindowMS  int `toml:"window_ms"`
}

type IPRateLimiterConfig struct {
	Register IPRateLimiterSection `toml:"register"`
	Login    IPRateLimiterSection `toml:"login"`
	Logout   IPRateLimiterSection `toml:"logout"`
	Profile  IPRateLimiterSection `toml:"profile"`
}

type Config struct {
	IPRateLimiter IPRateLimiterConfig `toml:"ip_rate_limiter"`
	DBPath        string              `toml:"dbpath"`
	Debug         bool                `toml:"debug"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Convert section → actual IPRateLimiter instance using builder API.
func BuildIPRLimiter(sec IPRateLimiterSection) *protect.IPRateLimiter {
	return protect.NewIPRateLimiter().
		MaxRequests(sec.Threshold).
		Window(time.Duration(sec.WindowMS) * time.Millisecond).
		Build()
}
