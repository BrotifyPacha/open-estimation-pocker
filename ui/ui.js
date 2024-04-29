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
