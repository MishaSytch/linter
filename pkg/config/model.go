package config

type Config struct {
	Loggers        []string     `yaml:"loggers"`
	SensitiveWords []string     `yaml:"sensitive_words"`
	Output         OutputConfig `yaml:"output"`
}

type OutputConfig struct {
	ShowInConsole   bool `yaml:"show_in_console" default:"false"`
	ShowSuggestions bool `yaml:"show_suggestions" default:"false"`
	ErrorsAggregate bool `yaml:"errors_aggregate" default:"false"`
	TestConfig      bool `yaml:"is_test" default:"true"`
}
