package test_case

import (
	"fmt"
	"log"
	"log/slog"
)

type LogMsg string

const (
	hardPrefix       = "Error: "
	hardMsgWithEmoji = hardPrefix + easyMsgSpecialChars

	typedLowerMsg LogMsg = "Upper case start"
	sensitiveData        = "access_token = 12345"
	cyrillicMixed        = "log level: очень высокий"
)

func testLogWhenHardThenOk() {
	log.Print("Fatal" + " error!" + easyMsgSpecialChars) // want "log message should start with a lowercase letter" "log message shouldn`t contain special characters or emojis"
	log.Print(typedLowerMsg)                             // want "log message should start with a lowercase letter"
	log.Print(hardMsgWithEmoji)                          // want "log message should start with a lowercase letter" "log message shouldn`t contain special characters or emojis"
	log.Print(cyrillicMixed)                             // want "log message should be in English"
	log.Printf("critical error: %s", sensitiveData)      // want "log message contains potentially sensitive data: token"
}

func testSlogWhenHardThenOk() {
	msgFromFunc := func() string { return "Dynamic Error" }
	slog.Error(msgFromFunc())
}

func testWhenAdvancedNestedThenOk() {
	log.Print(fmt.Sprintf("User password is: %s", "12345")) // want "log message should start with a lowercase letter" "log message contains potentially sensitive data: password"
}

func testWhenEdgeCasesThenOk() {
	log.Print("123 system is up")
	log.Print("...loading configuration")
	log.Print("v1.0.4 deployed")
	log.Print("")
}
