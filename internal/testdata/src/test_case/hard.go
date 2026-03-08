package test_case

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"log/slog"
)

type CustomStatus string

const StatusFailed CustomStatus = "FAILED_SYSTEM_ERROR"

type Request struct {
	ID    string
	Token string
}

func (r *Request) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("id", r.ID)
	enc.AddString("token", r.Token) // want "log message contains potentially sensitive data: token" "log message contains potentially sensitive data: token"
	return nil
}

type Credentials struct {
	Password string
}

func (c Credentials) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("password", c.Password) // want "log message contains potentially sensitive data: password" "log message contains potentially sensitive data: password"
	return nil
}

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

func testZapExtremeCases() {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	_ = &Request{ID: "123", Token: "secret-abc"} // здесь должно быть нормально, так как мы не выводим никуда эту структуру

	req := &Request{ID: "123", Token: "secret-abc"} // want "log message contains potentially sensitive data: secret" "log message contains potentially sensitive data: token" "log message contains potentially sensitive data: secret" "log message contains potentially sensitive data: token"
	logger.Info("processing request", zap.Object("req", req))

	logger.With(zap.String("context", "auth")).
		Named("security_module").
		Error("Auth failure") // want "log message should start with a lowercase letter"

	msg := "Alert: " + string(StatusFailed) // want "log message should start with a lowercase letter"
	sugar.Warn(msg)

	creds := Credentials{Password: "admin"} // want "log message contains potentially sensitive data: password" "log message contains potentially sensitive data: password"
	logger.Info("user attempt", zap.Inline(creds))

	logger.Debug("trace", customField("session_token", "val")) // want "log message contains potentially sensitive data: token"
}

func customField(key, val string) zap.Field {
	return zap.String(key, val)
}
