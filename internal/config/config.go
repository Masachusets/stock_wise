package config

import "flag"

type Config struct {
	DB   string
	Port int
	Log  string
}

func FromFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.DB, "db", "postgres://stockwise:stockwise@localhost:5432/stockwise", "PostgreSQL connection string")
	flag.IntVar(&cfg.Port, "port", 8080, "HTTP port")
	flag.StringVar(&cfg.Log, "log", "info", "Log level: debug, info, warn, error")
	flag.Parse()
	return cfg
}
