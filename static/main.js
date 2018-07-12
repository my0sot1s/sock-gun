function getClientId() {
  return window["client_id"];
}
function setClientId(clientId) {
  window.client_id = clientId;
}
window.onload = function() {
  var conn;
  var msg = document.getElementById("msg");
  var log = document.getElementById("log");

  function appendLog(item) {
    var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
    log.appendChild(item);
    if (doScroll) {
      log.scrollTop = log.scrollHeight - log.clientHeight;
    }
  }

  document.getElementById("form").onsubmit = function() {
    if (!conn) {
      return false;
    }
    if (!msg.value) {
      return false;
    }
    console.log({ data: msg.value, client_id: getClientId() });
    conn.send(JSON.stringify({ data: msg.value, client_id: getClientId() }));
    msg.value = "";
    return false;
  };

  if (!window["WebSocket"]) {
    var item = document.createElement("div");
    item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    appendLog(item);
    return;
  }
  conn = new WebSocket("ws://" + document.location.host + "/ws");
  conn.onclose = function(evt) {
    var item = document.createElement("div");
    item.innerHTML = "<b>Connection closed.</b>";
    appendLog(item);
  };
  conn.onmessage = function(evt) {
    let message = {};
    try {
      message = JSON.parse(evt.data);
    } catch (e) {
      message.data = evt.data;
    }
    switch (message.type) {
      case "upsert_client_id":
        setClientId(message.client_id);
        break;
      case "message_text":
        text = message.data.split("\n");
        var item = document.createElement("div");
        item.innerText = text[0];
        appendLog(item);
      default:
        return;
    }
  };
};
