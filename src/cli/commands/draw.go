package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func drawCommand(options *rootOptions) *cobra.Command {
	var offer bool
	var accept bool
	command := &cobra.Command{
		Use:   "draw --offer|--accept",
		Short: "Offer or accept a draw",
		RunE: func(command *cobra.Command, _ []string) error {
			if offer == accept {
				return fmt.Errorf("specify exactly one of --offer or --accept")
			}
			session, err := options.currentSession(command.Context())
			if err != nil {
				return err
			}
			response, raw, err := options.client().Draw(command.Context(), session.GameID, session.PlayerID)
			if err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printRaw(writer, raw)
			}
			if accept {
				if err := writeString(writer, "Draw accepted.\n"); err != nil {
					return err
				}
			} else {
				if err := writeString(writer, "Draw offered.\n"); err != nil {
					return err
				}
			}
			if response.GameState != nil {
				return writef(writer, "Status: %s\n", response.GameState.Status)
			}
			return nil
		},
	}
	command.Flags().BoolVar(&offer, "offer", false, "offer a draw")
	command.Flags().BoolVar(&accept, "accept", false, "accept a draw")
	return command
}
