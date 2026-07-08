package handlers

import (
	"net/http"
	"strconv"

	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

// GetValidMoves returns the valid destination squares for a piece at (row,col).
func (h *Handlers) GetValidMoves(w http.ResponseWriter, r *http.Request) {
	g, err := h.store.LoadGame(gameID(r))
	if err != nil {
		writeError(w, errorStatus(err), err.Error())
		return
	}

	rowStr := r.URL.Query().Get("row")
	colStr := r.URL.Query().Get("col")
	row, err := strconv.Atoi(rowStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid row parameter")
		return
	}
	col, err := strconv.Atoi(colStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid col parameter")
		return
	}

	pos := board.Position{Row: row, Col: col}
	if !pos.IsValid() {
		writeError(w, http.StatusBadRequest, "position out of bounds")
		return
	}

	moves, err := h.validator.GetValidMoves(g, pos)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if moves == nil {
		moves = []board.Position{}
	}

	writeJSON(w, http.StatusOK, dto.ValidMovesResponse{
		Success: true,
		Moves:   moves,
	})
}
