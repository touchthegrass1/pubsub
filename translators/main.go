package main

import (
	"log"
	"os"
	"translators/src"
)

func main() {
	log.Printf("Server started\n")

	data := make(chan string, 10000)
	defer close(data)
	router := src.NewRouter(data)

	log.Fatal(router.Run(":" + os.Getenv("TRANSLATOR_PORT")))
}
