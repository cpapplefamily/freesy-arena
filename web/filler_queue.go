// Author: justin.d.fischer@icloud.com
//
// Web handlers for the filler queue panel

package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Team254/cheesy-arena/game"

	"github.com/Team254/cheesy-arena/model"
)

// *****begin practice match filler line handler*****
func (web *Web) addPracticeMatchGetHandler(w http.ResponseWriter, r *http.Request) {
	template, err := web.parseFiles(
		"templates/filler_queue.html",
		"templates/base.html",
	)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	matches, err := web.arena.Database.GetMatchesByType(model.Practice, true)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	teamMap := make(map[int]*model.Team)
	for _, team := range teams {
		teamMap[team.Id] = &team
	}

	// Default time: now + 15min
	editMatchTime := time.Now().Add(15 * time.Minute).Format("2006-01-02T15:04")

	// If a matchId is in the query string, get the match time from DB
	if matchIdStr := r.URL.Query().Get("matchId"); matchIdStr != "" {
		if matchId, err := strconv.Atoi(matchIdStr); err == nil {
			if match, err := web.arena.Database.GetMatchById(matchId); err == nil && match.Type == model.Practice {
				editMatchTime = match.Time.Format("2006-01-02T15:04")
			}
		}
	}

	data := struct {
		EventSettings  *model.EventSettings
		Matches        []model.Match
		Teams          map[int]*model.Team
		MatchTimeValue string
	}{
		EventSettings:  web.arena.EventSettings,
		Matches:        matches,
		Teams:          teamMap,
		MatchTimeValue: editMatchTime,
	}

	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
	}
}

func (web *Web) addPracticeMatchPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	red := [3]int{
		parseTeam(r.FormValue("red1")),
		parseTeam(r.FormValue("red2")),
		parseTeam(r.FormValue("red3")),
	}
	blue := [3]int{
		parseTeam(r.FormValue("blue1")),
		parseTeam(r.FormValue("blue2")),
		parseTeam(r.FormValue("blue3")),
	}

	// Fetch existing practice matches to determine next match number
	matches, err := web.arena.Database.GetMatchesByType(model.Practice, true)
	if err != nil {
		http.Error(w, "Failed to fetch matches", http.StatusInternalServerError)
		return
	}

	nextTypeOrder := 1
	nextMatchTime := time.Now()
	if len(matches) > 0 {
		last := matches[len(matches)-1]
		nextTypeOrder = last.TypeOrder + 1
		nextMatchTime = last.Time.Add(15 * time.Minute)
	}

	match := &model.Match{
		Type:       model.Practice,
		TypeOrder:  nextTypeOrder,
		Time:       nextMatchTime,
		Red1:       red[0],
		Red2:       red[1],
		Red3:       red[2],
		Blue1:      blue[0],
		Blue2:      blue[1],
		Blue3:      blue[2],
		Status:     game.MatchScheduled,
		LongName:   fmt.Sprintf("Practice %d", nextTypeOrder),
		ShortName:  fmt.Sprintf("P%d", nextTypeOrder),
		NameDetail: "",
	}

	if err := web.arena.Database.CreateMatch(match); err != nil {
		http.Error(w, "Failed to save match", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/panel/freezy/add_practice_match", http.StatusSeeOther)
}

func (web *Web) editPracticeMatchHandler(w http.ResponseWriter, r *http.Request) {
	matchId, _ := strconv.Atoi(r.FormValue("matchId"))
	match, err := web.arena.Database.GetMatchById(matchId)
	if err != nil || match.Type != model.Practice {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}

	// Parse team numbers
	match.Red1 = parseTeam(r.FormValue("red1"))
	match.Red2 = parseTeam(r.FormValue("red2"))
	match.Red3 = parseTeam(r.FormValue("red3"))
	match.Blue1 = parseTeam(r.FormValue("blue1"))
	match.Blue2 = parseTeam(r.FormValue("blue2"))
	match.Blue3 = parseTeam(r.FormValue("blue3"))

	// Parse and set new match time
	timeStr := r.FormValue("matchTime")
	if timeStr != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04", timeStr) // HTML datetime-local format
		if err != nil {
			http.Error(w, "Invalid match time format", http.StatusBadRequest)
			return
		}
		match.Time = parsedTime
	}

	// Save match
	err = web.arena.Database.UpdateMatch(match)
	if err != nil {
		http.Error(w, "Failed to update match", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/panel/freezy/add_practice_match", http.StatusSeeOther)
}

func parseTeam(val string) int {
	num, _ := strconv.Atoi(val)
	return num
}
