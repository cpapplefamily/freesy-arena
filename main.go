// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

// Go version 1.22 or newer is required.
//go:build go1.22

package main

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/web"
	"log"
	"github.com/Team254/cheesy-arena/mqtt"
	"os"
	"os/signal"
	"syscall"
)

const eventDbPath = "./event.db"
const httpPort = 8080

// Main entry point for the application.
func main() {
	// Create and start the MQTT broker
	mqttBroker := mqtt.NewBroker()
	mqttBroker.Start()

	// Set up signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	
	// Create the arena	
	arena, err := field.NewArena(eventDbPath)
	if err != nil {
		log.Fatalln("Error during startup: ", err)
	}

	// Start the web server in a separate goroutine.
	web := web.NewWeb(arena)
	go web.ServeWebInterface(httpPort)

	// Run the arena state machine in the main thread.
	arena.Run()

}
