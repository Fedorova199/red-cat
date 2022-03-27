package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

const (
	defaultServerAddress = ":8080"
	defaultBaseURL       = "http://localhost:8080"
)

var defaultConfig = Config{
	ServerAddress: defaultServerAddress,
	BaseURL:       defaultBaseURL,
}

func NewConfig() (Config, error) {
	conf := defaultConfig
	conf.parseFlags()
	conf.parseEnvVars()
	err := conf.Validate()
	return conf, err
}

func (conf *Config) parseFlags() {

	flag.StringVar(&conf.ServerAddress, "a", defaultServerAddress, "network address the server listens on")
	flag.StringVar(&conf.BaseURL, "b", defaultBaseURL, "resulting base URL")
	flag.StringVar(&conf.FileStoragePath, "f", "test.txt", "storage file")
	flag.StringVar(&conf.DatabaseDSN, "d", "", `database dsn (default "")`)
	flag.Parse()

}

func (conf *Config) parseEnvVars() {
	sa := os.Getenv("SERVER_ADDRESS")
	if sa != "" {
		conf.ServerAddress = sa
	}

	bu := os.Getenv("BASE_URL")
	if bu != "" {
		conf.BaseURL = bu
	}

	fsp := os.Getenv("FILE_STORAGE_PATH")
	if fsp != "" {

		conf.FileStoragePath = fsp
	}

	dd, ok := os.LookupEnv("DATABASE_DSN")
	if ok {

		conf.DatabaseDSN = dd
	}
}

func (conf *Config) Validate() error {

	conf.ServerAddress = strings.TrimSpace(conf.ServerAddress)
	conf.BaseURL = strings.TrimSpace(conf.BaseURL)
	conf.FileStoragePath = strings.TrimSpace(conf.FileStoragePath)

	return nil
}
