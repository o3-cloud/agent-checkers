// Package display renders terminal-friendly checkers output.
package display

import (
	"strconv"
	"strings"
)

const separator = "  +---+---+---+---+---+---+---+---+"

// RenderBoard returns an ASCII board with unicode checkers pieces.
func RenderBoard(board [][]interface{}) string {
	var builder strings.Builder
	builder.WriteString("    0   1   2   3   4   5   6   7\n")
	builder.WriteString(separator)
	builder.WriteByte('\n')
	for row := 0; row < 8; row++ {
		builder.WriteString(strconv.Itoa(row))
		builder.WriteString(" |")
		for col := 0; col < 8; col++ {
			builder.WriteString(" ")
			builder.WriteString(pieceSymbol(pieceAt(board, row, col)))
			builder.WriteString(" |")
		}
		builder.WriteByte('\n')
		builder.WriteString(separator)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func pieceAt(board [][]interface{}, row, col int) map[string]interface{} {
	if row >= len(board) || col >= len(board[row]) {
		return nil
	}
	piece, ok := board[row][col].(map[string]interface{})
	if !ok {
		return nil
	}
	return piece
}

func pieceSymbol(piece map[string]interface{}) string {
	if piece == nil {
		return " "
	}
	isKing, _ := piece["is_king"].(bool)
	if isKing {
		return "♛"
	}
	color, _ := piece["color"].(string)
	switch color {
	case "black":
		return "●"
	case "red":
		return "○"
	default:
		return " "
	}
}
