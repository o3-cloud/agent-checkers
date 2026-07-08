package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func vsCommand(options *rootOptions) *cobra.Command {
	var aiName string
	var playerName string
	command := &cobra.Command{
		Use:   "vs --ai <name> --name <name>",
		Short: "Start a game against an AI player",
		RunE: func(command *cobra.Command, _ []string) error {
			if aiName == "" {
				return fmt.Errorf("--ai is required")
			}
			if playerName == "" {
				playerName = "Player"
			}
			apiClient := options.client()
			playerResponse, playerRaw, err := apiClient.CreateGame(command.Context(), playerName, "human")
			if err != nil {
				return err
			}
			aiResponse, aiRaw, err := apiClient.JoinGame(command.Context(), playerResponse.GameID, aiName, "ai")
			if err != nil {
				return err
			}
			if err := options.saveSession(command.Context(), playerResponse); err != nil {
				return err
			}
			writer := options.writer(command)
			if options.json {
				return printJSON(writer, map[string]any{
					"success": true,
					"player":  jsonRaw(playerRaw),
					"ai":      jsonRaw(aiRaw),
				})
			}
			if err := writef(writer, "Game created: %s\n", playerResponse.GameID); err != nil {
				return err
			}
			if err := writef(writer, "You are: %s\n", playerResponse.Player.Color); err != nil {
				return err
			}
			if err := writef(writer, "AI opponent: %s\n", aiResponse.Player.Name); err != nil {
				return err
			}
			if aiResponse.GameState != nil {
				if err := writef(writer, "Status: %s\n", aiResponse.GameState.Status); err != nil {
					return err
				}
				return writef(writer, "Turn: %s\n", aiResponse.GameState.CurrentTurn)
			}
			return nil
		},
	}
	command.Flags().StringVar(&aiName, "ai", "", "AI player name")
	command.Flags().StringVar(&playerName, "name", "", "human player name")
	return command
}

func jsonRaw(raw []byte) any {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	return value
}
