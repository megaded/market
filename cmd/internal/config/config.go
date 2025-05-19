package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const key = "1434535454545435435435435"

type Config struct {
	Address              string `env:"RUN_ADDRESS"`
	DbConnString         string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
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
	key := flag.String("k", key, "key")
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
	if c.SecretKey == "" {
		c.SecretKey = *key
	}
}
