package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func joinCommand(options *rootOptions) *cobra.Command {
	var name string
	command := &cobra.Command{
		Use:   "join <game-id> --name <name>",
		Short: "Join an existing game",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			response, raw, err := options.client().JoinGame(command.Context(), args[0], name, "human")
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
			if err := writef(writer, "Joined game: %s\n", response.GameID); err != nil {
				return err
			}
			if err := writef(writer, "You are: %s\n", response.Player.Color); err != nil {
				return err
			}
			if response.GameState != nil {
				if err := writef(writer, "Status: %s\n", response.GameState.Status); err != nil {
					return err
				}
				if response.GameState.RedPlayer != nil {
					if err := writef(writer, "Red player: %s\n", response.GameState.RedPlayer.Name); err != nil {
						return err
					}
				}
				if response.GameState.BlackPlayer != nil {
					if err := writef(writer, "Black player: %s\n", response.GameState.BlackPlayer.Name); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	command.Flags().StringVar(&name, "name", "", "player name")
	return command
}
