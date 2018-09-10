// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the pit rankings display.

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"net/http"
)

// Renders the pit display which shows scrolling rankings.
func (web *Web) pitDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	if !web.enforceDisplayConfiguration(w, r, nil) {
		return
	}

	template, err := web.parseFiles("templates/pit_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "pit_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the pit display, used only to force reloads remotely.
func (web *Web) pitDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsReader(w, r) {
		return
	}

	display, err := field.DisplayFromUrl(r.URL.Path, r.URL.Query())
	if err != nil {
		handleWebErr(w, err)
		return
	}
	web.arena.RegisterDisplay(display)
	defer web.arena.MarkDisplayDisconnected(display)

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(web.arena.DisplayConfigurationNotifier, web.arena.ReloadDisplaysNotifier)
}
