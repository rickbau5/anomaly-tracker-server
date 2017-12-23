package tracker

import (
	"flag"
	"time"

	"github.com/go-sql-driver/mysql"
)

// The AppConfig struct holds all configuration for the app.
type AppConfig struct {
	ListenAddr   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	MySQLHost string
	MySQLPort string
	MySQLUser string
	MySQLPass string
}

func (ac *AppConfig) BuildMySQLConfig() mysql.Config {
	return mysql.Config{
		User:         ac.MySQLUser,
		Passwd:       ac.MySQLPass,
		Addr:         ac.MySQLHost,
		Net:          "tcp",
		Timeout:      10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// InitConfig populates and returns a config by initializing and parsing runtime flags.
func InitConfig() AppConfig {
	conf := AppConfig{}
	initFlags(&conf)
	flag.Parse()
	return conf
}

func initFlags(config *AppConfig) {
	flag.StringVar(&config.ListenAddr, "listen-addr", ":8081", "listen address for the server")
	flag.DurationVar(&config.ReadTimeout, "read-timeout", 5*time.Second, "http read timeout of the server ")
	flag.DurationVar(&config.WriteTimeout, "write-timeout", 5*time.Second, "http write timeout of the server ")
	flag.DurationVar(&config.IdleTimeout, "idle-timeout", 15*time.Second, "http idle timeout of the server ")

	flag.StringVar(&config.MySQLHost, "mysql-host", "192.168.99.100:3306", "MySQL database host address")
	flag.StringVar(&config.MySQLUser, "mysql-user", "devadmin", "MySQL user")
	flag.StringVar(&config.MySQLPass, "mysql-pass", "devadmin", "MySQL password")
}
