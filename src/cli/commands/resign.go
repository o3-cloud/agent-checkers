package commands

import "github.com/spf13/cobra"

func resignCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "resign",
		Short: "Resign the current game",
		RunE: func(command *cobra.Command, _ []string) error {
			session, err := options.currentSession(command.Context())
			if err != nil {
				return err
			}
			response, raw, err := options.client().Resign(command.Context(), session.GameID, session.PlayerID)
			if err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printRaw(writer, raw)
			}
			if err := writeString(writer, "Resigned.\n"); err != nil {
				return err
			}
			if response.GameState != nil {
				return writef(writer, "Status: %s\n", response.GameState.Status)
			}
			return nil
		},
	}
}
