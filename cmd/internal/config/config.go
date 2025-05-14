package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	Address              string `env:"ADDRESS"`
	DbConnString         string `env:"ADDRESS"`
	AccrualSystemAddress string `env:"ADDRESS"`
}

func GetConfig() Config {
	config := Config{}
	setEnvParam(&config)
	setCmdParam(&config)
	return config
}

func setEnvParam(c *Config) {
	env.Parse(c)
}

func setCmdParam(c *Config) {
	address := flag.String("a", "", "address")
	dBConnString := flag.String("d", "", "db conn string")
	accrualSystemAddress := flag.String("i", "", "accrual system address")
	flag.Parse()
	if c.Address == "" {
		c.Address = *address
	}
	if c.DbConnString == "" {
		c.DbConnString = *dBConnString
	}
	if c.AccrualSystemAddress == "" {
		c.AccrualSystemAddress = *accrualSystemAddress
	}
}
