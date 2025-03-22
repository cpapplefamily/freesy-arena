// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the field monitor display showing robot connection status.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
)

// Renders the field monitor display.
func (web *Web) openQueueKioskHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if !web.enforceDisplayConfiguration(w, r, map[string]string{}) {
		return
	}

	template, err := web.parseFiles("templates/templates/open_queue.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		Teams            []model.Team
	}{web.arena.EventSettings, teams}
	err = template.ExecuteTemplate(w, "open_queue.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the field monitor display client to receive status updates.
func (web *Web) openQueueKioskWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplay(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchTimingNotifier, display.Notifier, web.arena.ArenaStatusNotifier,
		web.arena.EventStatusNotifier, web.arena.RealtimeScoreNotifier, web.arena.MatchTimeNotifier,
		web.arena.MatchLoadNotifier, web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		if command == "substituteTeams" {
			args := struct {
				Red1  int
				Red2  int
				Red3  int
				Blue1 int
				Blue2 int
				Blue3 int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
			err = web.arena.SubstituteTeams(args.Red1, args.Red2, args.Red3, args.Blue1, args.Blue2, args.Blue3)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}
		}
	}
}
