package main

import (
	"log"

	"github.com/montrosesoftware/tarpon/pkg/agent"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/config"
	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

func main() {
	log.Printf("starting tarpon...")

	config := config.ParseConfig()
	store := messaging.NewRoomStore()
	broker := broker.NewBroker()
	logger := logging.NewLogrusLogger(&config.Logging)
	server := server.NewRoomServer(store, agent.PeerHandler(broker, logger), logger)

	server.Listen(config.Server.Host, config.Server.Port)

}
