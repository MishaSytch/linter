package model

//FIXME в дальнейшем нужно перенести в конфиг

// SensitiveKeywords Чувствительные слова
var SensitiveKeywords = []string{
	"password",
	"token",
	"api_key",
	"secret",
}

//FIXME в дальнейшем нужно перенести в конфиг

// LoggerList Список анализируемых логеров
var LoggerList = []string{
	"log",
	"slog",
	"zap",
	"logger",
}
