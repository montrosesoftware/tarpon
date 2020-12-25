package main

import (
	"log"

	"github.com/montrosesoftware/tarpon/pkg/agent"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/config"
	"github.com/montrosesoftware/tarpon/pkg/instrumentation"
	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

func main() {
	log.Printf("starting tarpon...")

	config := config.ParseConfig()
	logger := logging.NewLogrusLogger(&config.Logging)
	store := messaging.NewRoomStore()
	broker := broker.NewBroker(logger)
	server := server.NewRoomServer(store, agent.PeerHandler(broker, logger), logger)

	instrumentation := instrumentation.NewPrometheusInstrumentation()
	server.EnableMetrics(instrumentation.MetricsHandler())

	server.Listen(config.Server.Host, config.Server.Port)

}
