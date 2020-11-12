package config

import (
	"github.com/jinzhu/configor"
)

const (
	ModeRelease = "release"
	ModeDebug   = "debug"
)

//Config contains app settings
type Config struct {
	Mode string
	DB   struct {
		Type     string
		Settings string
	}
	Backend struct {
		Listen string
	}
}

func app() *Config {
	cfg := Config{}
	configor.Load(&cfg, ".env.yml")
	return &cfg
}

//App config
var App = app()
