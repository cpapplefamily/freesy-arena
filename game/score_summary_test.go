// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreSummaryDetermineMatchStatus(t *testing.T) {
	redScoreSummary := &ScoreSummary{Score: 10}
	blueScoreSummary := &ScoreSummary{Score: 10}
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.Score = 11
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.Score = 12
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.Score = 12
	redScoreSummary.NumOpponentMajorFouls = 11
	redScoreSummary.AutoPoints = 11
	redScoreSummary.BargePoints = 11
	blueScoreSummary.NumOpponentMajorFouls = 10
	blueScoreSummary.AutoPoints = 10
	blueScoreSummary.BargePoints = 10
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.NumOpponentMajorFouls = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.NumOpponentMajorFouls = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.AutoPoints = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.AutoPoints = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.BargePoints = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.BargePoints = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
}
