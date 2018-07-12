package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	clientID string
}

// Events is a type of event
type Events struct {
	events   map[string]string
	clientID string
}

func (e *Events) setEvents(clientID string, evts map[string]string) {
	e.clientID = clientID
	rootEvents := e.events
	for key, value := range evts {
		rootEvents[key] = value
	}
	e.events = rootEvents
}

func (e *Events) removeEvents(clientID string, evts map[string]string) {
	e.clientID = clientID
	rootEvents := e.events
	for key, _ := range evts {
		delete(rootEvents, key)
	}
	e.events = rootEvents
}

func (c *Client) messageDefine(t string) map[string]string {
	message := make(map[string]string)
	switch t {
	case "upsert_client_id":
		message["type"] = "upsert_client_id"
		message["client_id"] = c.clientID
		break
	case "message_text":
		message["type"] = "message_text"
	}
	return message

}

// generateClientId is create a new client Id
func (c *Client) generateClientID() string {
	if c.clientID != "" {
		return c.clientID
	}
	c.clientID = "ws" + randStringRunes(20)
	return c.clientID
}

func (c *Client) handleSetConnectionID() {
	c.generateClientID()
	if msgData, err := json.Marshal(c.messageDefine("upsert_client_id")); c.clientID != "" {
		// update clientId to client
		if err != nil {
			fmt.Printf("%v", err)
			return
		}
		c.send <- msgData
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, raw, err := c.conn.ReadMessage()
		fmt.Printf("%s", raw)
		fmt.Println()
		message := make(map[string]string)
		json.Unmarshal(raw, &message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if clientID := message["client_id"]; clientID == "" {
			c.handleSetConnectionID()
		} else {
			message := c.messageDefine("message_text")
			message["data"] = "flow done"
			if bin, er := json.Marshal(message); er == nil {
				c.send <- bin
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
