package handlers

import (
	"net/http"

	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

// MakeMove validates, executes, and persists a move.
func (h *Handlers) MakeMove(w http.ResponseWriter, r *http.Request) {
	var request dto.MoveRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	g, err := h.store.LoadGame(gameID(r))
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	if err := g.MakeMove(request.PlayerID, request.From, request.To); err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}
	if err := h.validator.ValidateMove(g, request.From, request.To); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	captured, err := h.validator.ExecuteMove(g, request.From, request.To)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(captured) == 0 || !h.validator.HasCapturesForPiece(g, request.To) {
		g.SwitchTurn()
	}

	if over, result := g.IsGameOver(); over {
		g.EndGame(result)
	}

	if err := h.store.SaveGame(g); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	move := g.Moves[len(g.Moves)-1]
	writeJSON(w, http.StatusOK, dto.MoveResponse{
		Success:   true,
		Move:      ptr(dto.NewMoveResponse(move)),
		GameState: dto.NewGameState(g),
	})
}

// GetMoveHistory returns all moves made in a game.
func (h *Handlers) GetMoveHistory(w http.ResponseWriter, r *http.Request) {
	moves, err := h.store.GetMoveHistory(gameID(r))
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.MoveHistoryResponse{
		Success: true,
		Moves:   dto.NewMoveHistoryResponse(moves),
	})
}

func ptr[T any](value T) *T {
	return &value
}
