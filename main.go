// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

// Go version 1.22 or newer is required.
//go:build go1.22

package main

import (
	"log"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/web"
)

const eventDbPath = "./event.db"
const httpPort = 8080
var version = "dev"

// Main entry point for the application.
func main() {
	log.Printf("App version: %s\n", version)
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
