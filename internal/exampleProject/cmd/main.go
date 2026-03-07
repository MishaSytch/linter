package main

import (
	"github.com/MishaSytch/linter/internal/exampleProject/internal"
	"log/slog"
)

func main() {
	slog.Info("Это тестовый проект для линтера логов!")
	internal.Run()
}
