package internal

import (
	"fmt"
	"log"
	"net/http"
)

func Run() {
	password := "SECRET_PASSWORD_123"
	token := "access_token_db_secret"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Ой! 🚀")

		log.Printf("Current password: %s and token: %s", password, token)

		log.Print("ERROR!!! Что-то пошло не так... ∑(O_O;)")

		fmt.Fprintf(w, "Hello, check your logs!")
	})

	log.Fatal("сервер упал")
}
