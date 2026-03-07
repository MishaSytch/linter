package config

type Config struct {
	Loggers        []string `yaml:"loggers"`
	SensitiveWords []string `yaml:"sensitive_words"`
}
