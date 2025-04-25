package wspad

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Pad struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
	content []byte
	path    string
}

var pads = make(map[string]*Pad)
var padsMu sync.Mutex

func getOrCreatePad(padName string) *Pad {
	padsMu.Lock()
	defer padsMu.Unlock()

	if pad, ok := pads[padName]; ok {
		return pad
	}

	filePath := filepath.Join(DataDir, padName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading pad file %s: %v", padName, err)
		}
		content = []byte{}
	}

	pad := &Pad{
		clients: make(map[*websocket.Conn]bool),
		content: content,
		path:    filePath,
	}
	pads[padName] = pad
	return pad
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	padName := strings.TrimPrefix(r.URL.Path, "/_ws/")
	if padName == "" || strings.Contains(padName, "/") || strings.Contains(padName, "..") {
		log.Printf("Invalid pad name requested via WebSocket: %s", padName)
		http.NotFound(w, r)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for pad %s: %v", padName, err)
		return
	}
	defer conn.Close()

	pad := getOrCreatePad(padName)

	pad.mu.Lock()
	pad.clients[conn] = true
	initialContent := pad.content
	pad.mu.Unlock()

	log.Printf("Client connected to pad: %s", padName)

	if err := conn.WriteMessage(websocket.TextMessage, initialContent); err != nil {
		log.Printf("Error sending initial content to client for pad %s: %v", padName, err)
		pad.removeClient(conn)
		return
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client disconnected unexpectedly from pad %s: %v", padName, err)
			} else {
				log.Printf("Error reading message from client for pad %s: %v", padName, err)
			}
			pad.removeClient(conn)
			break
		}

		if messageType == websocket.TextMessage {
			pad.handleMessage(conn, message)
		}
	}
}

func (p *Pad) handleMessage(sender *websocket.Conn, message []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.content = message

	go func(path string, content []byte) {
		if err := os.WriteFile(path, content, 0644); err != nil {
			log.Printf("Error writing pad file %s: %v", filepath.Base(path), err)
		}
	}(p.path, append([]byte(nil), message...))

	for client := range p.clients {
		if client != sender {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error broadcasting message to client for pad %s: %v", filepath.Base(p.path), err)
			}
		}
	}
}

func (p *Pad) removeClient(conn *websocket.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.clients[conn]; ok {
		delete(p.clients, conn)
		conn.Close()
		log.Printf("Client removed from pad: %s", filepath.Base(p.path))

		// if len(p.clients) == 0 {
		// 	padsMu.Lock()
		// 	delete(pads, filepath.Base(p.path))
		// 	padsMu.Unlock()
		// 	log.Printf("Pad %s closed as no clients are connected.", filepath.Base(p.path))
		// }
	}
}
