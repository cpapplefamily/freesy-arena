// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function(data) {
  $("#matchName").text(data.Match.LongName);
  if (alliance === "red") {
    $(".team-1").text(data.Match.Red1);
    $(".team-2").text(data.Match.Red2);
    $(".team-3").text(data.Match.Red3);
  } else {
    $(".team-1").text(data.Match.Blue1);
    $(".team-2").text(data.Match.Blue2);
    $(".team-3").text(data.Match.Blue3);
  }
};

// Handles a websocket message to update the match status.
const handleMatchTime = function(data) {
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
const handleRealtimeScore = function(data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;
    $(`#leaveStatus${i1}>.value`).text(score.LeaveStatuses[i] ? "Yes" : "No");
    $(`#leaveStatus${i1}`).attr("data-value", score.LeaveStatuses[i]);
    $(`#parkTeam${i1}`).attr("data-value", score.EndgameStatuses[i] === 1);
    $(`#stageSide0Team${i1}`).attr("data-value", score.EndgameStatuses[i] === 2);
    $(`#stageSide1Team${i1}`).attr("data-value", score.EndgameStatuses[i] === 3);
    $(`#stageSide2Team${i1}`).attr("data-value", score.EndgameStatuses[i] === 4);
    $(`#stageSide${i}Microphone`).attr("data-value", score.MicrophoneStatuses[i]);
    $(`#stageSide${i}Trap`).attr("data-value", score.TrapStatuses[i]);
    $("#endgameStatus" + i1 + ">.value").text(getEndgameStatusText(score.EndgameStatuses[i]));
    $("#endgameStatus" + i1).attr("data-value", score.EndgameStatuses[i]);
  }

  //Some Diagnostis
  $("#currentScore").text("Current Score: " + realtimeScore.ScoreSummary.Score);
  //$("#currentAmpificationCount").text("Banked Amp Notes: " + score.AmpSpeaker.BankedAmpNotes);
  $("#processorCount").text("Processed Algae Count: " + score.AmpSpeaker.ProcessedAlgae);
  //$("#ampCount").text("Amp Total Count: " + ( score.AmpSpeaker.TeleopAmpNotes + 
  //                                            score.AmpSpeaker.AutoAmpNotes));
  //$("#teleopAmpCount").text(score.TeleopAmpNotes);
  //$("#autoAmpCount").text(score.AutoAmpNotes);
  $("#speakerCount").text("Speaker Total Count: " + ( score.AmpSpeaker.AutoSpeakerNotes + 
                                                      score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes +
                                                      score.AmpSpeaker.TeleopAmplifiedSpeakerNotes));
  //$("#autoSpeakerCount").text(score.AutoSpeakerNotes);
  //$("#teleopSpeakerCountNotAmplified").text(score.TeleopSpeakerNotesNotAmplified);
  //$("#teleopSpeakerCountAmplified").text(score.TeleopSpeakerNotesAmplified);
  //$("#trapCount1").text((score.TrapNotes));
  //$("#trapCount").text("Trap Count: " + (score.TrapNotes));

  
  $(`#coopertitionStatus>.value`).text(score.AmpSpeaker.CoopActivated ? "Cooperation Enabled" : "Cooperation");
  $("#coopertitionStatus").attr("data-value", score.AmpSpeaker.CoopActivated);
  $(`#amplificationActive>.value`).text(realtimeScore.AmplifiedTimePostWindow ? "Amplification Active" : "Amplification");
  $("#amplificationActive").attr("data-value", realtimeScore.AmplifiedTimePostWindow);
  $("#amplificationActive").css("background-color", !(realtimeScore.AmplifiedTimeRemainingSec > 0) && realtimeScore.AmplifiedTimePostWindow? "yellow" : "");
  $("#amplificationActive").css("color", !(realtimeScore.AmplifiedTimeRemainingSec > 0) && realtimeScore.AmplifiedTimePostWindow  ? "black" : "");

  $("#processedAlgae").text(score.AmpSpeaker.ProcessedAlgae);
  $("#netAlgae").text(score.AmpSpeaker.NetAlgae);
  $("#autoLvL1_0").text(score.Grid.AutoLvL1Count[0]);
  $("#teliopLvL1_0").text(score.Grid.TeliopLvL1Count[0]);
  $("#autoLvL1_1").text(score.Grid.AutoLvL1Count[1]);
  $("#teliopLvL1_1").text(score.Grid.TeliopLvL1Count[1]);
  //$("#autoSpeakerNotes").text(score.AmpSpeaker.AutoSpeakerNotes);
  //$("#autoSpeakerNotes").text(score.AmpSpeaker.AutoSpeakerNotes);
  $("#teleopAmplifiedSpeakerNotes").text(score.AmpSpeaker.TeleopAmplifiedSpeakerNotes);
  $("#teleopUnamplifiedSpeakerNotes").text(score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes);
  $("#autoAmpNotes").text(score.AmpSpeaker.AutoAmpNotes);
  $("#teleopAmpNotes").text(score.AmpSpeaker.TeleopAmpNotes);
  $("#bankedAmpNotes").text(score.AmpSpeaker.BankedAmpNotes);

  for (let i = 0; i < 4; i++) {
    for (let j = 0; j < 12; j++) {
      $(`#gridAutoScoringRow${i}Node${j}`).attr("data-value", score.Grid.AutoScoring[i][j]);
      $(`#gridNodeStatesRow${i}Node${j}`).children().each(function() {
        const element = $(this);
        element.attr("data-value", element.attr("data-node-state") === score.Grid.Nodes[i][j].toString());
      });
    }
  }
};

// Returns the display text corresponding to the given integer endgame status value.
const getEndgameStatusText = function(level) {
  switch (level) {
    case 1:
      return "Park";
    case 2:
      return "Shallow";
    case 3:
      return "Deep";
    default:
      return "None";
  }
};

// Handles an element click and sends the appropriate websocket message.
const handleClick = function(command, teamPosition = 0, stageIndex = 0, gridRow = 0, gridNode = 0, nodeState = 0) {
  websocket.send(command, {TeamPosition: teamPosition, StageIndex: stageIndex, GridRow: gridRow, GridNode: gridNode, NodeState: nodeState});
};

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);
  $("#manualScore").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });
});

// Set initial visibility state
let bool = true;

// Function to toggle visibility of the panel
function toggleEstopPanel() {
    const panel = document.querySelector('.eStops');
    if (bool) {
        panel.classList.remove('hidden'); // Hide the panel
    } else {
        panel.classList.add('hidden'); // Show the panel
    }
    bool = !bool; // Toggle the boolean state
}

// Attach event listener to the button
document.getElementById('toggleButton').addEventListener('click', toggleEstopPanel);
