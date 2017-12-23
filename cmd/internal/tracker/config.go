package tracker

import (
	"flag"
	"time"
)

// The ServerConfig struct holds all configuration for the app.
type ServerConfig struct {
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// InitConfig populates and returns a config by initializing and parsing runtime flags.
func InitConfig() ServerConfig {
	conf := ServerConfig{}
	initFlags(&conf)
	flag.Parse()
	return conf
}

func initFlags(config *ServerConfig) {
	flag.StringVar(&config.ListenAddr, "listen-addr", ":8081", "listen address for the server")
	flag.DurationVar(&config.ReadTimeout, "read-timeout", 5*time.Second, "http read timeout of the server ")
	flag.DurationVar(&config.WriteTimeout, "write-timeout", 5*time.Second, "http write timeout of the server ")
	flag.DurationVar(&config.IdleTimeout, "idle-timeout", 15*time.Second, "http idle timeout of the server ")
}
