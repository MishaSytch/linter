package config

type Config struct {
	Loggers        []string       `yaml:"loggers"`
	SensitiveRules SensitiveRules `yaml:"sensitive_rules"`
	Output         OutputConfig   `yaml:"output"`
}

type SensitiveRules struct {
	Patterns       []SensitivePattern `yaml:"patterns"`
	SensitiveWords []string           `yaml:"words"`
}

type SensitivePattern struct {
	Name  string `yaml:"name"`
	Regex string `yaml:"regex"`
}

type OutputConfig struct {
	ShowInConsole   bool `yaml:"show_in_console" default:"false"`
	ShowSuggestions bool `yaml:"show_suggestions" default:"false"`
	ErrorsAggregate bool `yaml:"errors_aggregate" default:"false"`
	TestRun         bool `yaml:"is_test" default:"true"`
}
