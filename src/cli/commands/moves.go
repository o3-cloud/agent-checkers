package commands

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stackable-specs/agent-checkers/internal/app/board"
)

func movesCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "moves",
		Short: "List legal moves for the saved player",
		RunE: func(command *cobra.Command, _ []string) error {
			session, err := options.currentSession(command.Context())
			if err != nil {
				return err
			}
			apiClient := options.client()
			gameResponse, _, err := apiClient.GetGame(command.Context(), session.GameID)
			if err != nil {
				return err
			}
			color, err := playerColor(gameResponse.GameState, session.PlayerID)
			if err != nil {
				return err
			}
			moves, err := collectLegalMoves(command.Context(), apiClient, session.GameID, gameResponse.GameState.Board, color)
			if err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printJSON(writer, map[string]any{
					"success": true,
					"moves":   moves,
				})
			}
			if len(moves) == 0 {
				return writeString(writer, "No legal moves.\n")
			}
			for _, move := range moves {
				if err := writef(writer, "(%d,%d) -> (%d,%d)\n", move.From.Row, move.From.Col, move.To.Row, move.To.Col); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

type validMovesClient interface {
	ValidMoves(ctx context.Context, gameID string, pos board.Position) ([]board.Position, []byte, error)
}

func collectLegalMoves(ctx context.Context, apiClient validMovesClient, gameID string, squares [][]interface{}, color string) ([]legalMove, error) {
	var moves []legalMove
	for row := 0; row < len(squares); row++ {
		for col := 0; col < len(squares[row]); col++ {
			if boardPieceColor(squares[row][col]) != color {
				continue
			}
			from := board.Position{Row: row, Col: col}
			destinations, _, err := apiClient.ValidMoves(ctx, gameID, from)
			if err != nil {
				return nil, err
			}
			for _, to := range destinations {
				moves = append(moves, legalMove{From: from, To: to})
			}
		}
	}
	return moves, nil
}
