package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v2"
)

const filename = "tarpon.yaml"

type Config struct {
	Logging Logging
	Server  Server
}

type Logging struct {
	Level string `yaml:"level" env:"TARPON_LOGGING_LEVEL" env-description:"Log level. One of trace, debug, info, warn or error" env-default:"info"`
}

type Server struct {
	Host string `yaml:"host" env:"TARPON_HOST" env-description:"Server host. All by default" env-default:""`
	Port string `yaml:"port" env:"TARPON_PORT" env-description:"Server post." env-default:"5000"`
}

func ParseConfig() Config {
	var cfg Config

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("%s not found, reading config from environment", filename)
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("error reading config: %v", err)
		}
	} else {
		log.Printf("%s found, reading config from %s and environment", filename, filename)
		if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
			log.Fatalf("error reading config: %v", err)
		}
	}

	yaml, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Fatalf("error while printing config: %v", err)
	}
	log.Printf("tarpon config:\n%s\n", string(yaml))

	return cfg
}
