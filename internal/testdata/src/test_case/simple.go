package test_case

import (
	"log"
	"log/slog"
)

const (
	easyMsgLowerCaseRules = "Starting server..."
	easyMsgEnglish        = "tití me preguntó"
	easyMsgSpecialChars   = "es normal – 🫡"
	easyMsgSensitive      = "password for database root user: postgres"
)

var easyOk = []string{
	"test is ok",
	"",
	"123 - = _ + asd 4985 BigWord",
}

func testLogWhenEasyThenOk() {
	log.Print(easyMsgLowerCaseRules) // want "log message should start with a lowercase letter"
	log.Print(easyMsgEnglish)        // want "log message should be in English"
	log.Print(easyMsgSpecialChars)   // want "log message shouldn`t contain special characters or emojis"
	log.Printf(easyMsgSensitive)     // want "log message contains potentially sensitive data: password"

	for _, msg := range easyOk {
		log.Print(msg)
	}
}

func testSlogWhenEasyThenOk() {
	slog.Error(easyMsgLowerCaseRules) // want "log message should start with a lowercase letter"
	slog.Error(easyMsgEnglish)        // want "log message should be in English"
	slog.Error(easyMsgSpecialChars)   // want "log message shouldn`t contain special characters or emojis"
	slog.Error(easyMsgSensitive)      // want "log message contains potentially sensitive data: password"

	for _, msg := range easyOk {
		slog.Info(msg)
	}
}
