package main

import (
	"log"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalln("error listening on port 3333")
	}
}