package main

import (
	"log"
	"net/http"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

func main() {
	server := messaging.RoomServer{}
	if err := http.ListenAndServe(":5000", &server); err != nil {
		log.Fatalf("can't listen on port 5000 %v", err)
	}
}