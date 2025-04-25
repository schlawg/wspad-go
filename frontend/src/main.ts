let ws: WebSocket;
let editor: HTMLTextAreaElement;
let pingTimeout: number;
let connectTimeout: number | undefined;
let connectTimer = 300 + Math.random() * 200;

window.onload = () => {
  editor = document.querySelector('#text') as HTMLTextAreaElement;
  editor.oninput = (e: any) => send(e.target.value);
  connect();
};

function connect() {
  connectTimeout = undefined;
  ws = new WebSocket(`wss://${window.location.host}/_ws${window.location.pathname}`);
  ws.binaryType = 'arraybuffer';
  ws.onopen = e => {
    editor.disabled = false;
  };
  ws.onmessage = onMessage;
  ws.onclose = onClose;
  ws.onerror = function (e) {
    console.error(e);
    editor.disabled = true;
    ws.close();
  };
}

function schedulePing() {
  clearTimeout(pingTimeout);
  if (connectTimeout) return;
  if (ws.readyState < WebSocket.CLOSING) pingTimeout = setTimeout(send, 10000);
  else if (ws.readyState !== WebSocket.CLOSED) ws.close();
}

function clear() {
  clearTimeout(pingTimeout);
  clearTimeout(connectTimeout);
  connectTimeout = undefined;
}

function send(text?: string) {
  ws.send(text ?? new Int32Array(1));
  schedulePing();
}

function onClose(e: any) {
  editor.disabled = true;
  clear();
  connectTimeout = setTimeout(connect, connectTimer);
  connectTimer *= 2;
  if (connectTimer > 10000) connectTimer = 300 + Math.random() * 200;
}

function onMessage(e: any) {
  if (typeof e.data !== 'string') return; // pong
  editor.value = e.data;
  schedulePing();
}
