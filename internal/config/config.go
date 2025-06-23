package config

import (
	"os"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config is the main configuration struct.
type Config struct {
	Env       string          `koanf:"env"`
	Tarantool TarantoolConfig `koanf:"tarantool"`
	HTTP      HTTPConfig      `koanf:"http"`
}

// TarantoolConfig is the configuration for the Tarantool instance.
type TarantoolConfig struct {
	Host     string        `koanf:"host"`
	Port     int           `koanf:"port"`
	User     string        `koanf:"user"`
	Password string        `koanf:"password"`
	Timeout  time.Duration `koanf:"timeout"`
	KVSpace  string        `koanf:"kv_space"`
	KVIndex  string        `koanf:"kv_index"`
}

// HTTPConfig is the configuration for the HTTP server.
type HTTPConfig struct {
	Port       int           `koanf:"port"`
	Timeout    time.Duration `koanf:"timeout"`
	KVBasePath string        `koanf:"kv_base_path"`
}

// MustLoad returns the configuration loaded from the environment, in case of
// error it panics.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

// Load returns the configuration loaded from the environment.
func Load() (*Config, error) {
	configPath := os.Getenv("KV_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/local.yml"
	}

	k := koanf.New(":")
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, err
	}

	c := &Config{}
	if err := k.Unmarshal("", &c); err != nil {
		return nil, err
	}

	return c, nil
}
