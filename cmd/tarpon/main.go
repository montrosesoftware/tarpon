package main

import (
	"log"
	"net/http"

	"github.com/montrosesoftware/tarpon/pkg/agent"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

func main() {
	store := messaging.NewRoomStore()
	broker := broker.NewBroker()
	logger := logging.NewLogrusLogger()
	server := server.NewRoomServer(store, agent.PeerHandler(broker, logger))
	log.Printf("server working...")
	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("can't listen on port 5000 %v", err)
	}
}
