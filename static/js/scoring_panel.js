// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var alliance;

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  $("#matchName").text(data.MatchType + " " + data.Match.DisplayName);
  if (alliance === "red") {
    $("#team1").text(data.Match.Red1);
    $("#team2").text(data.Match.Red2);
    $("#team3").text(data.Match.Red3);
  } else {
    $("#team1").text(data.Match.Blue1);
    $("#team2").text(data.Match.Blue2);
    $("#team3").text(data.Match.Blue3);
  }
};

// Handles a websocket message to update the match status.
var handleMatchTime = function(data) {
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      // Pre-match message state is set in handleRealtimeScore().
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
      break;
    case "POST_MATCH":
      $("#postMatchMessage").hide();
      $("#commitMatchScore").css("display", "flex");
      break;
    default:
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
  }
};

// Handles a websocket message to update the realtime scoring fields.
var handleRealtimeScore = function(data) {
  var realtimeScore1;
  var realtimeScore2;
  if (alliance === "red") {
    realtimeScore1 = data.Red;
    realtimeScore2 = data.Blue;
  } else {
    realtimeScore1 = data.Blue;
    realtimeScore2 = data.Red;
  }
  var score1 = realtimeScore1.Score;
  var score2 = realtimeScore2.Score;

  //Group One Score
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    $("#taxiStatus" + i1 + ">.value").text(score1.TaxiStatuses[i] ? "Yes" : "No");
    $("#taxiStatus" + i1).attr("data-value", score1.TaxiStatuses[i]);
    $("#endgameStatus" + i1 + ">.value").text(getEndgameStatusText(score1.EndgameStatuses[i]));
    $("#endgameStatus" + i1).attr("data-value", score1.EndgameStatuses[i]);
    $("#autoCargoLower").text(score1.AutoCargoLower[0]);
    $("#autoCargoUpper").text(score1.AutoCargoUpper[0]);
    $("#teleopCargoLower").text(score1.TeleopCargoLower[0]);
    $("#teleopCargoUpper").text(score1.TeleopCargoUpper[0]);
  }

  //Group Two Score 
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    $("#taxiStatus2" + i1 + ">.value").text(score2.TaxiStatuses[i] ? "Yes" : "No");
    $("#taxiStatus2" + i1).attr("data-value", score2.TaxiStatuses[i]);
    $("#endgameStatus2" + i1 + ">.value").text(getEndgameStatusText(score2.EndgameStatuses[i]));
    $("#endgameStatus2" + i1).attr("data-value", score2.EndgameStatuses[i]);
    $("#autoCargoLower2").text(score2.AutoCargoLower[0]);
    $("#autoCargoUpper2").text(score2.AutoCargoUpper[0]);
    $("#teleopCargoLower2").text(score2.TeleopCargoLower[0]);
    $("#teleopCargoUpper2").text(score2.TeleopCargoUpper[0]);
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  websocket.send(String.fromCharCode(event.keyCode));
};

// Handles an element click and sends the appropriate websocket message.
var handleClick = function(shortcut) {
  websocket.send(shortcut);
};

// Sends a websocket message to indicate that the score for this alliance is ready.
var commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

// Returns the display text corresponding to the given integer endgame status value.
var getEndgameStatusText = function(level) {
  switch (level) {
    case 1:
      return "Low";
    case 2:
      return "Mid";
    case 3:
      return "High";
    case 4:
      return "Traversal";
    default:
      return "None";
  }
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });

  $(document).keypress(handleKeyPress);
});
