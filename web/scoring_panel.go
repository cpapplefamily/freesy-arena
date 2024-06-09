// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"time"

)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	alliance := r.PathValue("alliance")
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled bool
		Alliance     string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	alliance := r.PathValue("alliance")
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	var realtimeScore **field.RealtimeScore
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier)

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
		score := &(*realtimeScore).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else {
			args := struct {
				TeamPosition int
				StageIndex   int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			switch command {
				case "leave":
					if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
						score.LeaveStatuses[args.TeamPosition-1] = !score.LeaveStatuses[args.TeamPosition-1]
						scoreChanged = true
					}
				case "onStage":
					if args.TeamPosition >= 1 && args.TeamPosition <= 3 && args.StageIndex >= 0 && args.StageIndex <= 2 {
						endgameStatus := game.EndgameStatus(args.StageIndex + 2)
						if score.EndgameStatuses[args.TeamPosition-1] == endgameStatus {
							score.EndgameStatuses[args.TeamPosition-1] = game.EndgameNone
						} else {
							score.EndgameStatuses[args.TeamPosition-1] = endgameStatus
						}
						scoreChanged = true
					}
				case "park":
					if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
						if score.EndgameStatuses[args.TeamPosition-1] == game.EndgameParked {
							score.EndgameStatuses[args.TeamPosition-1] = game.EndgameNone
						} else {
							score.EndgameStatuses[args.TeamPosition-1] = game.EndgameParked
						}
						scoreChanged = true
					}
				case "microphone":
					log.Printf("Microphone Pressed")
					if args.StageIndex >= 0 && args.StageIndex <= 2 {
						score.MicrophoneStatuses[args.StageIndex] = !score.MicrophoneStatuses[args.StageIndex]
						scoreChanged = true
					}
				case "trap":
					if args.StageIndex >= 0 && args.StageIndex <= 2 {
						score.TrapStatuses[args.StageIndex] = !score.TrapStatuses[args.StageIndex]
						scoreChanged = true
					}
				case "S":
					var _matchStartTime = web.arena.MatchStartTime
					var _currentTime = time.Now()
					score.AmpSpeaker.UpdateState(	score.AmpSpeaker.AutoAmpNotes + 
													score.AmpSpeaker.TeleopAmpNotes,	
													
													score.AmpSpeaker.AutoSpeakerNotes + 
													score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes +
													score.AmpSpeaker.TeleopAmplifiedSpeakerNotes	+	1, 
													
													false,
													
													false, 
													
													_matchStartTime,
													
													_currentTime)
						log.Printf("Speaker Pressed")
						scoreChanged = true
				case "O":
					score.AmpSpeaker.AutoSpeakerNotes = score.AmpSpeaker.AutoSpeakerNotes+1
					log.Printf("O Pressed")
					scoreChanged = true
				case "o":
					if score.AmpSpeaker.AutoSpeakerNotes > 0 {
						score.AmpSpeaker.AutoSpeakerNotes = score.AmpSpeaker.AutoSpeakerNotes-1
					}
					log.Printf("o Pressed")
					scoreChanged = true
				case "P":
					score.AmpSpeaker.TeleopAmplifiedSpeakerNotes = score.AmpSpeaker.TeleopAmplifiedSpeakerNotes+1
					log.Printf("P Pressed")
					scoreChanged = true
				case "p":
					if score.AmpSpeaker.TeleopAmplifiedSpeakerNotes > 0{
						score.AmpSpeaker.TeleopAmplifiedSpeakerNotes = score.AmpSpeaker.TeleopAmplifiedSpeakerNotes-1
					}
					log.Printf("p Pressed")
					scoreChanged = true
				case "I":
					score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes = score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes+1
					log.Printf("I Pressed")
					scoreChanged = true
				case "i":
					if score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes > 0{
						score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes = score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes-1
					}
					log.Printf("i Pressed")
					scoreChanged = true
				case "A":
					var _matchStartTime = web.arena.MatchStartTime
					var _currentTime = time.Now()
					score.AmpSpeaker.UpdateState(	score.AmpSpeaker.AutoAmpNotes + 
													score.AmpSpeaker.TeleopAmpNotes + 1, 
													
													score.AmpSpeaker.AutoSpeakerNotes + 
													score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes +
													score.AmpSpeaker.TeleopAmplifiedSpeakerNotes,
													
													false,
														
													false, 
													
													_matchStartTime,
														
													_currentTime)
					log.Printf("Amp Pressed")
					scoreChanged = true
				case "U":
					score.AmpSpeaker.AutoAmpNotes = score.AmpSpeaker.AutoAmpNotes+1
					if score.AmpSpeaker.BankedAmpNotes < 2{
						score.AmpSpeaker.BankedAmpNotes = score.AmpSpeaker.BankedAmpNotes+1
					}
					log.Printf("U Pressed")
					scoreChanged = true
				case "u":
					if score.AmpSpeaker.AutoAmpNotes > 0 {
						score.AmpSpeaker.AutoAmpNotes = score.AmpSpeaker.AutoAmpNotes-1
					}
					log.Printf("u Pressed")
					scoreChanged = true
				case "Y":
					score.AmpSpeaker.TeleopAmpNotes = score.AmpSpeaker.TeleopAmpNotes+1
					if score.AmpSpeaker.BankedAmpNotes < 2{
						score.AmpSpeaker.BankedAmpNotes = score.AmpSpeaker.BankedAmpNotes+1
					}
					log.Printf("Y Pressed")
					scoreChanged = true
				case "y":
					if score.AmpSpeaker.TeleopAmpNotes > 0 {
						score.AmpSpeaker.TeleopAmpNotes = score.AmpSpeaker.TeleopAmpNotes-1
					}
					log.Printf("y Pressed")
					scoreChanged = true
				case "B":
					if score.AmpSpeaker.BankedAmpNotes < 2{
						score.AmpSpeaker.BankedAmpNotes = score.AmpSpeaker.BankedAmpNotes+1
					}
					log.Printf("B Pressed")
					scoreChanged = true
				case "b":
					if score.AmpSpeaker.BankedAmpNotes > 0 {
						score.AmpSpeaker.BankedAmpNotes = score.AmpSpeaker.BankedAmpNotes-1
					}
					log.Printf("b Pressed")
					scoreChanged = true
				case "amplifyButton":
					var _matchStartTime = web.arena.MatchStartTime
					var _currentTime = time.Now()
					score.AmpSpeaker.UpdateState(	score.AmpSpeaker.AutoAmpNotes + 
													score.AmpSpeaker.TeleopAmpNotes, 
					
													score.AmpSpeaker.AutoSpeakerNotes + 
													score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes +
													score.AmpSpeaker.TeleopAmplifiedSpeakerNotes,
													
													true,
													
													false, 
													
													_matchStartTime,
													
													_currentTime)
					log.Printf("amplifyButton Pressed")
					scoreChanged = true
				case "coopButton":
					var _matchStartTime = web.arena.MatchStartTime
					var _currentTime = time.Now()
					score.AmpSpeaker.UpdateState(	score.AmpSpeaker.AutoAmpNotes + 
													score.AmpSpeaker.TeleopAmpNotes, 

													score.AmpSpeaker.AutoSpeakerNotes + 
													score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes +
													score.AmpSpeaker.TeleopAmplifiedSpeakerNotes,

													false,
														
													true, 
														
													_matchStartTime,
														
													_currentTime)
					log.Printf("coopButton Pressed")
					scoreChanged = true
			}

			if scoreChanged {
				web.arena.RealtimeScoreNotifier.Notify()
			}
		}
	}
}
