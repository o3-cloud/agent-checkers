package websocket

import (
	"encoding/json"
	"sync"
)

// Hub manages WebSocket rooms keyed by game ID.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

// Room manages all connected clients for one game.
type Room struct {
	gameID  string
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewHub creates an empty WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

// RegisterClient adds a client to its game room.
func (h *Hub) RegisterClient(client *Client) {
	if client == nil {
		return
	}
	room := h.room(client.gameID)
	room.add(client)
}

// UnregisterClient removes a client from its game room.
func (h *Hub) UnregisterClient(client *Client) {
	if client == nil {
		return
	}

	h.mu.RLock()
	room := h.rooms[client.gameID]
	h.mu.RUnlock()
	if room == nil {
		return
	}

	room.remove(client)
	if room.count() != 0 {
		return
	}

	h.mu.Lock()
	if room.count() == 0 && h.rooms[client.gameID] == room {
		delete(h.rooms, client.gameID)
	}
	h.mu.Unlock()
}

// BroadcastToGame sends a raw message to every client connected to a game.
func (h *Hub) BroadcastToGame(gameID string, message []byte) {
	h.mu.RLock()
	room := h.rooms[gameID]
	h.mu.RUnlock()
	if room == nil {
		return
	}
	room.broadcast(message)
}

// BroadcastEvent marshals and sends an event to every client connected to a game.
func (h *Hub) BroadcastEvent(gameID string, event Event) {
	message, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.BroadcastToGame(gameID, message)
}

func (h *Hub) room(gameID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := h.rooms[gameID]
	if room == nil {
		room = &Room{
			gameID:  gameID,
			clients: make(map[*Client]struct{}),
		}
		h.rooms[gameID] = room
	}
	return room
}

func (r *Room) add(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = struct{}{}
}

func (r *Room) remove(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.clients[client]; !ok {
		return
	}
	delete(r.clients, client)
	close(client.send)
}

func (r *Room) broadcast(message []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for client := range r.clients {
		select {
		case client.send <- message:
		default:
			delete(r.clients, client)
			close(client.send)
		}
	}
}

func (r *Room) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}
