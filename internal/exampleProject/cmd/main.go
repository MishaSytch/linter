package main

import (
	"exampleProject/internal"
	"log/slog"
)

func main() {
	slog.Info("Это тестовый проект для линтера логов!")
	internal.Run()
}
