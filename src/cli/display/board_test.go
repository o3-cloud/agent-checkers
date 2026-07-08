package display

import (
	"strings"
	"testing"
)

func TestRenderBoard(t *testing.T) {
	t.Parallel()

	board := make([][]interface{}, 8)
	for row := range board {
		board[row] = make([]interface{}, 8)
	}
	board[0][1] = map[string]interface{}{"color": "black", "is_king": false}
	board[1][0] = map[string]interface{}{"color": "red", "is_king": false}
	board[7][6] = map[string]interface{}{"color": "red", "is_king": true}

	got := RenderBoard(board)
	checks := []string{
		"    0   1   2   3   4   5   6   7\n",
		"0 |   | ● |",
		"1 | ○ |",
		"7 |   |   |   |   |   |   | ♛ |",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("RenderBoard() missing %q in:\n%s", check, got)
		}
	}
}

func TestRenderBoardHandlesShortBoard(t *testing.T) {
	t.Parallel()

	got := RenderBoard(nil)
	if !strings.Contains(got, "7 |   |   |   |   |   |   |   |   |") {
		t.Fatalf("RenderBoard(nil) did not render empty rows:\n%s", got)
	}
}
