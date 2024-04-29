function ConnectViaWebsocket(roomId, username) {
    wsProtocol = 'ws'
    if (window.location.protocol.indexOf("https") != -1) {
        wsProtocol = 'wss'
    }

    socketUrl = `${wsProtocol}://${window.location.host}/${roomId}/ws?username=${username}`

    console.log(socketUrl);
    ws = new WebSocket(socketUrl);
    return ws
}

function addClass(view, cssClass) {
    view.className += ' ' + cssClass
}

function hide(view) {
    view.className += ' hidden'
}

function show(view) {
    view.className = view.className.replaceAll('hidden', '')
}

function arraysEqual(a, b) {
  if (a === b) return true;
  if (a == null || b == null) return false;
  if (a.length !== b.length) return false;

  // If you don't care about the order of the elements inside
  // the array, you should sort both arrays here.
  // Please note that calling sort on an array will modify that array.
  // you might want to clone your array first.

  for (var i = 0; i < a.length; ++i) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}
