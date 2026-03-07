package config

type Config struct {
	Loggers        []string     `yaml:"loggers"`
	SensitiveWords []string     `yaml:"sensitive_words"`
	Output         OutputConfig `yaml:"output"`
}

type OutputConfig struct {
	ShowInConsole   bool `yaml:"show_in_console"`
	ShowSuggestions bool `yaml:"show_suggestions"`
}
