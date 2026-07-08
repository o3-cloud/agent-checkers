// Package commands defines the agent-checkers CLI commands.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackable-specs/agent-checkers/internal/app/board"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
	"github.com/stackable-specs/agent-checkers/src/cli/client"
	clisession "github.com/stackable-specs/agent-checkers/src/cli/session"
)

const defaultServerURL = "http://localhost:8080"

type rootOptions struct {
	serverURL string
	json      bool
	out       io.Writer
}

type legalMove struct {
	From board.Position `json:"from"`
	To   board.Position `json:"to"`
}

// NewRootCommand creates the Cobra root command.
func NewRootCommand() *cobra.Command {
	options := &rootOptions{
		serverURL: defaultServerURL,
	}
	command := &cobra.Command{
		Use:   "agent-checkers-cli",
		Short: "Play agent-checkers from a terminal",
	}
	command.PersistentFlags().StringVar(&options.serverURL, "server", defaultServerURL, "REST API server URL")
	command.PersistentFlags().BoolVar(&options.json, "json", false, "output JSON")

	command.AddCommand(
		newCommand(options),
		joinCommand(options),
		boardCommand(options),
		moveCommand(options),
		movesCommand(options),
		resignCommand(options),
		drawCommand(options),
		vsCommand(options),
	)
	return command
}

func (o *rootOptions) writer(command *cobra.Command) io.Writer {
	if o.out != nil {
		return o.out
	}
	return command.OutOrStdout()
}

func (o *rootOptions) client() *client.Client {
	return client.New(o.serverURL)
}

func (o *rootOptions) currentSession(ctx context.Context) (*clisession.Session, error) {
	path, err := clisession.DefaultPath()
	if err != nil {
		return nil, err
	}
	session, err := clisession.Load(ctx, path)
	if err != nil {
		return nil, err
	}
	if session.ServerURL != "" && o.serverURL == defaultServerURL {
		o.serverURL = session.ServerURL
	}
	return session, nil
}

func (o *rootOptions) saveSession(ctx context.Context, response *dto.PlayerGameResponse) error {
	if response == nil || response.Player == nil {
		return fmt.Errorf("response did not include player")
	}
	path, err := clisession.DefaultPath()
	if err != nil {
		return err
	}
	token := ""
	if response.Session != nil {
		token = response.Session.Token
	}
	return clisession.Save(ctx, path, clisession.Session{
		GameID:       response.GameID,
		PlayerID:     response.Player.ID,
		PlayerName:   response.Player.Name,
		SessionToken: token,
		ServerURL:    o.serverURL,
	})
}

func printRaw(writer io.Writer, raw []byte) error {
	if _, err := writer.Write(raw); err != nil {
		return err
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		_, err := writer.Write([]byte("\n"))
		return err
	}
	return nil
}

func printJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeString(writer io.Writer, value string) error {
	_, err := io.WriteString(writer, value)
	return err
}

func writef(writer io.Writer, format string, args ...any) error {
	_, err := fmt.Fprintf(writer, format, args...)
	return err
}

func parsePosition(value string) (board.Position, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return board.Position{}, fmt.Errorf("position %q must use row,col format", value)
	}
	var pos board.Position
	if _, err := fmt.Sscanf(value, "%d,%d", &pos.Row, &pos.Col); err != nil {
		return board.Position{}, fmt.Errorf("parse position %q: %w", value, err)
	}
	if !pos.IsValid() {
		return board.Position{}, fmt.Errorf("position %q is out of bounds", value)
	}
	return pos, nil
}

func playerColor(state *dto.GameState, playerID string) (string, error) {
	if state == nil {
		return "", fmt.Errorf("game response did not include game_state")
	}
	if state.RedPlayer != nil && state.RedPlayer.ID == playerID {
		return "red", nil
	}
	if state.BlackPlayer != nil && state.BlackPlayer.ID == playerID {
		return "black", nil
	}
	return "", fmt.Errorf("saved player is not in game %s", state.ID)
}

func boardPieceColor(value interface{}) string {
	piece, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}
	color, _ := piece["color"].(string)
	return color
}
