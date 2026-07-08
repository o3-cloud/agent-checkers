package handlers

import (
	"net/http"

	"github.com/stackable-specs/agent-checkers/internal/app/game"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

// CreateGame creates a new waiting game and registers the first player.
func (h *Handlers) CreateGame(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateGameRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	g, player, err := h.lobby.RegisterPlayer(request.PlayerName, request.PlayerType)
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	session, err := h.createSession(player.ID, g.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.PlayerGameResponse{
		Success:   true,
		GameID:    g.ID,
		Player:    dto.NewPlayerResponse(player),
		Session:   dto.NewSessionResponse(session),
		GameState: dto.NewGameState(g),
	})
}

// JoinGame registers a second player for an existing waiting game.
func (h *Handlers) JoinGame(w http.ResponseWriter, r *http.Request) {
	var request dto.JoinGameRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	id := gameID(r)
	player, err := h.lobby.JoinGame(id, request.PlayerName, request.PlayerType)
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	g, err := h.store.LoadGame(id)
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	session, err := h.createSession(player.ID, g.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.broadcastGameStarted(g)

	writeJSON(w, http.StatusOK, dto.PlayerGameResponse{
		Success:   true,
		GameID:    g.ID,
		Player:    dto.NewPlayerResponse(player),
		Session:   dto.NewSessionResponse(session),
		GameState: dto.NewGameState(g),
	})
}

// GetGame returns the current state of a game.
func (h *Handlers) GetGame(w http.ResponseWriter, r *http.Request) {
	g, err := h.store.LoadGame(gameID(r))
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	})
}

// ResignGame ends a game by resignation.
func (h *Handlers) ResignGame(w http.ResponseWriter, r *http.Request) {
	var request dto.PlayerActionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	g, err := h.store.LoadGame(gameID(r))
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	if err := g.Resign(request.PlayerID); err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}
	if err := h.store.SaveGame(g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.broadcastGameEnded(g)

	writeJSON(w, http.StatusOK, dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	})
}

// OfferOrAcceptDraw creates a draw offer or accepts an existing opponent offer.
func (h *Handlers) OfferOrAcceptDraw(w http.ResponseWriter, r *http.Request) {
	var request dto.PlayerActionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	g, err := h.store.LoadGame(gameID(r))
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	if g.Result != nil && g.Result.DrawOffe != "" && g.Result.DrawOffe != request.PlayerID {
		err = g.AcceptDraw(request.PlayerID)
	} else {
		err = g.OfferDraw(request.PlayerID)
	}
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}
	if err := h.store.SaveGame(g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if g.Status == game.StatusDraw {
		h.broadcastGameEnded(g)
	}

	writeJSON(w, http.StatusOK, dto.GameResponse{
		Success:   true,
		GameID:    g.ID,
		GameState: dto.NewGameState(g),
	})
}

func (h *Handlers) createSession(playerID, gameID string) (*session.Session, error) {
	if h.sessions == nil {
		return nil, nil
	}
	return h.sessions.Create(playerID, gameID)
}
