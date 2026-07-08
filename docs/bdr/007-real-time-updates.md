# BDR-007: Real-Time Game Updates

## Status

Accepted

## Behavior

The system broadcasts game state changes to all connected clients in real-time via WebSocket, enabling instant visual updates without polling.

## Context

A turn-based game requires real-time updates so players can see when their opponent has moved. Polling the REST API creates unnecessary load and latency. WebSocket provides a persistent bidirectional connection for instant push notifications when game state changes.

## Acceptance Criteria

- AC-1: Players can connect to a WebSocket endpoint for their game
- AC-2: When a move is made, all connected clients receive a `move_made` event
- AC-3: When a game starts (second player joins), clients receive a `game_started` event
- AC-4: When a game ends, clients receive a `game_ended` event with winner information
- AC-5: WebSocket connection is authenticated via session token
- AC-6: Disconnected players can reconnect and receive the current game state
- AC-7: Heartbeat/ping-pong keeps connections alive
- AC-8: Multiple clients per player are supported (e.g., phone + desktop)

## Verification

### Scenario 1: Connect to game WebSocket

- **Given** a player has registered and received a session token
- **When** the player connects to `ws://host/api/v1/games/{id}/ws?token={session_token}`
- **Then** the connection is accepted
- **And** the player receives the current game state

### Scenario 2: Opponent move notification

- **Given** two players are connected to the game WebSocket
- **When** player A makes a move via REST API
- **Then** player B's WebSocket receives a `move_made` event within 100ms
- **And** the event contains `{ from, to, captured, board }`

### Scenario 3: Game start notification

- **Given** player A is waiting in the WebSocket
- **When** player B joins the game via REST API
- **Then** player A's WebSocket receives a `game_started` event
- **And** the event contains `{ game_id, players }`

### Scenario 4: Game end notification

- **Given** two players are connected
- **When** the game ends (win or draw)
- **Then** both players receive a `game_ended` event
- **And** the event contains `{ winner, reason }`

### Scenario 5: Reconnection after disconnect

- **Given** a player disconnects from WebSocket
- **When** the player reconnects with the same session token
- **Then** the connection is accepted
- **And** the player receives the current game state immediately

## Event Types

| Event | Direction | Payload |
|-------|-----------|---------|
| `game_started` | Server → Client | `{ game_id, players }` |
| `move_made` | Server → Client | `{ from, to, captured, board, turn }` |
| `turn_changed` | Server → Client | `{ current_player }` |
| `game_ended` | Server → Client | `{ winner, reason }` |
| `ping` | Server → Client | `{ timestamp }` |
| `pong` | Client → Server | `{ timestamp }` |

## Connection Lifecycle

```
1. Client connects with session token
2. Server validates token and retrieves player/game
3. Server joins client to game room (broadcast channel)
4. Server sends current game state
5. Server listens for game events and broadcasts to room
6. On disconnect, server removes client from room
```

## Technologies

| Component | Technology |
|-----------|------------|
| Protocol | WebSocket (RFC 6455) |
| Endpoint | `/api/v1/games/{id}/ws` |
| Authentication | Query param: `?token={session_token}` |
| Heartbeat | 30-second ping/pong |

## Interfaces

| Interface | Endpoint |
|-----------|----------|
| WebSocket | `GET /api/v1/games/{id}/ws` |

## Traceability

- Spec: `docs/spec.md` - Feature: WebSocket Protocol
- ADR: `docs/adr/007-interface-layers.md`