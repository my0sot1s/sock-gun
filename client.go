package main

import (
	"log"
	"net/http"
)

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// clientID := r.Header.Get("clientID")
	// fmt.Printf("Connect clientID:%s", clientID)
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256)}
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
