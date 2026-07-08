package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 65 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 1024
	sendBufferSize = 256
)

// Client represents one WebSocket connection to a game.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	gameID   string
	playerID string
	now      func() time.Time
}

func newClient(hub *Hub, conn *websocket.Conn, gameID, playerID string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		gameID:   gameID,
		playerID: playerID,
		now:      time.Now,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(c.now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(c.now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(c.now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			message, err := json.Marshal(Event{
				Type: EventTypePing,
				Payload: PingPayload{
					Timestamp: c.now(),
				},
			})
			if err != nil {
				return
			}
			_ = c.conn.SetWriteDeadline(c.now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) {
	var event Event
	if err := json.Unmarshal(message, &event); err != nil {
		return
	}
	if event.Type != EventTypePong {
		return
	}
	_ = c.conn.SetReadDeadline(c.now().Add(pongWait))
}

// SendEvent queues an event for delivery to this client.
func (c *Client) SendEvent(event Event) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}
	select {
	case c.send <- message:
		return nil
	default:
		c.hub.UnregisterClient(c)
		return nil
	}
}
