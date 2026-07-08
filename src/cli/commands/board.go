package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackable-specs/agent-checkers/src/cli/display"
)

func boardCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "board",
		Short: "Display the current board",
		RunE: func(command *cobra.Command, _ []string) error {
			session, err := options.currentSession(command.Context())
			if err != nil {
				return err
			}
			response, raw, err := options.client().GetGame(command.Context(), session.GameID)
			if err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printRaw(writer, raw)
			}
			if response.GameState == nil {
				return fmt.Errorf("game response did not include game_state")
			}
			if err := writeString(writer, display.RenderBoard(response.GameState.Board)); err != nil {
				return err
			}
			if err := writef(writer, "Status: %s\n", response.GameState.Status); err != nil {
				return err
			}
			return writef(writer, "Turn: %s\n", response.GameState.CurrentTurn)
		},
	}
}
