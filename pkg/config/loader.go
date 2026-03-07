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
			Loggers: model.LoggerList,
			SensitiveRules: SensitiveRules{
				Patterns: []SensitivePattern{
					{
						Name:  "Email Address",
						Regex: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
					},
				},
				SensitiveWords: model.SensitiveKeywords,
			},

			Output: OutputConfig{
				ShowInConsole:   false,
				ShowSuggestions: true,
				ErrorsAggregate: false,
				TestRun:         true,
			},
		}

		f, err := os.Open(path)
		if err != nil {
			slog.Warn("using default config", "path", path, "reason", "file not found")
			return
		}
		defer f.Close()

		if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
			slog.Error("failed to decode config", "error", err)
		}
	})
	return cfg
}
