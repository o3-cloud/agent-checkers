package commands

import (
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
)

func gamesCommand(options *rootOptions) *cobra.Command {
	var status string
	var playerID string
	command := &cobra.Command{
		Use:   "games",
		Short: "List active and waiting games",
		RunE: func(command *cobra.Command, _ []string) error {
			response, raw, err := options.client().ListGames(command.Context(), status, playerID)
			if err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printRaw(writer, raw)
			}
			if len(response.Games) == 0 {
				return writeString(writer, "No games found\n")
			}

			table := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
			if err := writeString(table, "GAME ID\tSTATUS\tTURN\tRED\tBLACK\n"); err != nil {
				return err
			}
			for _, g := range response.Games {
				if err := writef(table, "%s\t%s\t%s\t%s\t%s\n",
					g.GameID,
					g.Status,
					g.CurrentTurn,
					playerName(g.RedPlayer),
					playerName(g.BlackPlayer),
				); err != nil {
					return err
				}
			}
			return table.Flush()
		},
	}
	command.Flags().StringVar(&status, "status", "", "filter by status (waiting, active, completed, draw, all)")
	command.Flags().StringVar(&playerID, "player", "", "filter by player ID")
	return command
}

func playerName(player *dto.PlayerResponse) string {
	if player == nil {
		return "-"
	}
	return player.Name
}
