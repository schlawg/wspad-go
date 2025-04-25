package main

import (
	"log"
	"net/http"
	"os"
	"wspad/internal/wspad"
)

const addr = ":8080"

func main() {
	wspad.Init()

	if err := os.MkdirAll(wspad.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	fs := http.FileServer(http.Dir(wspad.StaticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", wspad.IndexHandler)
	http.HandleFunc("/_ws/", wspad.WSHandler)
	http.HandleFunc("/{padName}", wspad.PadHandler)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server on %s: %v", addr, err)
	}
}
