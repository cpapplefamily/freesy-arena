var handleStationTrip = function(data) {
  if (data.StationId !== station) return;

  const match = $("#match");
  match.removeClass("solid-orange blink-orange");

  if (!data.MatchInProgress) {
    if (data.EStopTripped) {
      match.addClass("solid-orange");
    } else if (data.AStopTripped) {
      match.addClass("blink-orange");
    }
  }
};