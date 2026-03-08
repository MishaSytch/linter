package test_case

import (
	"log"
	"log/slog"

	"go.uber.org/zap"
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

func testZapSugaredWhenEasyThenOk() {
	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()

	sugar.Info(easyMsgLowerCaseRules) // want "log message should start with a lowercase letter"
	sugar.Warn(easyMsgEnglish)        // want "log message should be in English"
	sugar.Error(easyMsgSpecialChars)  // want "log message shouldn`t contain special characters or emojis"

	sugar.Infow("User logged in", "password", "12345")  // want "log message should start with a lowercase letter" "log message contains potentially sensitive data: password"
	sugar.Errorf("Failed login for %s", "secret_token") // want "log message should start with a lowercase letter" "log message contains potentially sensitive data: token, secret"

	for _, msg := range easyOk {
		sugar.Info(msg)
	}
}

func testZapStructuredWhenEasyThenOk() {
	logger := zap.NewExample()
	defer logger.Sync()

	logger.Info(easyMsgLowerCaseRules) // want "log message should start with a lowercase letter"
	logger.Warn(easyMsgEnglish)        // want "log message should be in English"

	logger.Error("process failed",
		zap.String("api_key", "m1sha2"), // want "log message contains potentially sensitive data: api_key" "log message contains potentially sensitive data: api_key"
		zap.Int("attempts", 3),
	)

	for _, msg := range easyOk {
		logger.Info(msg)
	}
}

func testAllLoggerScenarios() {
	z, _ := zap.NewProduction()
	z.Info("Message") // want "log message should start with a lowercase letter"

	sl := slog.Default()
	sl.Warn("Bad message") // want "log message should start with a lowercase letter"

	getSugar().Error("Alert!") // want "log message should start with a lowercase letter"
	getSugar().Error("🫡")      // want "log message shouldn`t contain special characters or emojis"

	zap.L().Info("Global logger message") // want "log message should start with a lowercase letter"
	zap.S().Debug("Global sugar message") // want "log message should start with a lowercase letter"

	log.Print("Standard log message with password Mañana con 🫡") // want "log message should start with a lowercase letter" "log message contains potentially sensitive data: password" "log message shouldn`t contain special characters or emojis" "log message should be in English"
}

func getSugar() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}
