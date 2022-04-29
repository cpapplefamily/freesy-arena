// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
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

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	var realtimeScore1 **field.RealtimeScore
	var realtimeScore2 **field.RealtimeScore
	if alliance == "red" {
		realtimeScore1 = &web.arena.RedRealtimeScore
		realtimeScore2 = &web.arena.BlueRealtimeScore
	} else {
		realtimeScore1 = &web.arena.BlueRealtimeScore
		realtimeScore2 = &web.arena.RedRealtimeScore
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
		command, _, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		score1 := &(*realtimeScore1).CurrentScore
		score2 := &(*realtimeScore2).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if number, err := strconv.Atoi(command); err == nil && number >= 1 && number <= 6 {
			// Handle per-robot scoring fields.
			if number <= 3 {
				index := number - 1
				score1.TaxiStatuses[index] = !score1.TaxiStatuses[index]
				scoreChanged = true
			} else {
				index := number - 4
				score1.EndgameStatuses[index]++
				if score1.EndgameStatuses[index] == 5 {
					score1.EndgameStatuses[index] = 0
				}
				scoreChanged = true
			}
		} else if !web.arena.Plc.IsEnabled() {
			switch strings.ToUpper(command) {
				//Auto scoring is from a device running as Red Page
			case "RL":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score1.AutoCargoLower[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score1.TeleopCargoLower[:])
					}
				}

			case "RU":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score1.AutoCargoUpper[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score1.TeleopCargoUpper[:])
					}
				}
			case "BL":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score2.AutoCargoLower[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score2.TeleopCargoLower[:])
					}
				}
			case "BU":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score2.AutoCargoUpper[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score2.TeleopCargoUpper[:])
					}
				}
				//Group One
			case "Q":
				scoreChanged = decrementGoal(score1.AutoCargoUpper[:])
			case "A":
				scoreChanged = decrementGoal(score1.AutoCargoLower[:])
			case "W":
				scoreChanged = incrementGoal(score1.AutoCargoUpper[:])
			case "S":
				scoreChanged = incrementGoal(score1.AutoCargoLower[:])
			case "E":
				scoreChanged = decrementGoal(score1.TeleopCargoUpper[:])
			case "D":
				scoreChanged = decrementGoal(score1.TeleopCargoLower[:])
			case "R":
				scoreChanged = incrementGoal(score1.TeleopCargoUpper[:])
			case "F":
				scoreChanged = incrementGoal(score1.TeleopCargoLower[:])
				//Group tWO
			case "U":
				scoreChanged = decrementGoal(score2.AutoCargoUpper[:])
			case "H":
				scoreChanged = decrementGoal(score2.AutoCargoLower[:])
			case "I":
				scoreChanged = incrementGoal(score2.AutoCargoUpper[:])
			case "J":
				scoreChanged = incrementGoal(score2.AutoCargoLower[:])
			case "O":
				scoreChanged = decrementGoal(score2.TeleopCargoUpper[:])
			case "K":
				scoreChanged = decrementGoal(score2.TeleopCargoLower[:])
			case "P":
				scoreChanged = incrementGoal(score2.TeleopCargoUpper[:])
			case "L":
				scoreChanged = incrementGoal(score2.TeleopCargoLower[:])
			}

		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

// Increments the cargo count for the given goal.
func incrementGoal(goal []int) bool {
	// Use just the first hub quadrant for manual scoring.
	goal[0]++
	return true
}

// Decrements the cargo for the given goal.
func decrementGoal(goal []int) bool {
	// Use just the first hub quadrant for manual scoring.
	if goal[0] > 0 {
		goal[0]--
		return true
	}
	return false
}
