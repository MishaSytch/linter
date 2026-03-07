package config

import (
	"gopkg.in/yaml.v3"
	"linter/pkg/linter/model"
	"log/slog"
	"os"
	"sync"
)

var once sync.Once
var cfg *Config

func LoadConfig(path string) *Config {
	once.Do(func() {
		cfg = &Config{
			Loggers:        model.LoggerList,
			SensitiveWords: model.SensitiveKeywords,
		}

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return
		}

		f, err := os.Open(path)
		if err != nil {
			slog.Error("could not open config file (using default settings)", "path", path, "error", err)
			return
		}
		defer f.Close()

		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(cfg)
		if err != nil {
			slog.Error("failed to decode config yaml (using default settings)", "path", path, "error", err)
		}
		slog.Info("config loaded", "path", path)
	})

	return cfg
}
