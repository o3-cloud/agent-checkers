package commands

import "github.com/spf13/cobra"

func moveCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "move <from> <to>",
		Short: "Make a move, using row,col coordinates",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			from, err := parsePosition(args[0])
			if err != nil {
				return err
			}
			to, err := parsePosition(args[1])
			if err != nil {
				return err
			}
			session, err := options.currentSession(command.Context())
			if err != nil {
				return err
			}
			response, raw, err := options.client().MakeMove(command.Context(), session.GameID, session.PlayerID, from, to)
			if err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printRaw(writer, raw)
			}
			if err := writef(writer, "Move accepted: (%d,%d) -> (%d,%d)\n", from.Row, from.Col, to.Row, to.Col); err != nil {
				return err
			}
			if response.GameState != nil {
				if err := writef(writer, "Turn: %s\n", response.GameState.CurrentTurn); err != nil {
					return err
				}
				return writef(writer, "Status: %s\n", response.GameState.Status)
			}
			return nil
		},
	}
}
