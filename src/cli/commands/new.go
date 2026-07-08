package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCommand(options *rootOptions) *cobra.Command {
	var name string
	command := &cobra.Command{
		Use:   "new --name <name>",
		Short: "Create a new game",
		RunE: func(command *cobra.Command, _ []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			response, raw, err := options.client().CreateGame(command.Context(), name, "human")
			if err != nil {
				return err
			}
			if err := options.saveSession(command.Context(), response); err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printRaw(writer, raw)
			}
			if err := writef(writer, "Game created: %s\n", response.GameID); err != nil {
				return err
			}
			if err := writef(writer, "You are: %s\n", response.Player.Color); err != nil {
				return err
			}
			if err := writeString(writer, "Status: Waiting for opponent...\n"); err != nil {
				return err
			}
			return writef(writer, "Share: agent-checkers-cli join %s --name <name>\n", response.GameID)
		},
	}
	command.Flags().StringVar(&name, "name", "", "player name")
	return command
}
