package main

import (
	"log"
	"net/http"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

func main() {
	store := messaging.NewRoomStore()
	server := server.NewRoomServer(store)
	log.Printf("Server working...")
	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("can't listen on port 5000 %v", err)
	}
}
