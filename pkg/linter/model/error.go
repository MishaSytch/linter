package model

// Константы для сообщений об ошибках
const (
	// MsgLowerCaseRules – текст должен начинаться с прописной буквы
	MsgLowerCaseRules = "log message should start with a lowercase letter"

	// MsgEnglish – текст должен содержать только латиницу
	MsgEnglish = "log message should be in English"

	// MsgSpecialChars – текст не должен содержать специальный символы
	MsgSpecialChars = "log message shouldn`t contain special characters or emojis"

	// MsgSensitiveData – текст содержит чувствительные данные
	MsgSensitiveData = "log message contains potentially sensitive data: %s"
)
