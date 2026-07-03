package config

import "flag"

type Config struct {
	DB    string
	Port  int
	Log   string
	Debug bool
}

func FromFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.DB, "db", "postgres://stockwise:stockwise@localhost:5432/stockwise", "PostgreSQL connection string")
	flag.IntVar(&cfg.Port, "port", 8080, "HTTP port")
	flag.StringVar(&cfg.Log, "logFormat", "terminal", "Log Format: text, terminal, json")
	flag.BoolVar(&cfg.Debug, "debug", false, "Log request and respose bodies")
	flag.Parse()
	return cfg
}
