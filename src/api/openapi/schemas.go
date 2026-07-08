package openapi

// Schemas returns reusable component schemas for REST DTOs.
func Schemas() map[string]map[string]any {
	return map[string]map[string]any{
		"CreateGameRequest":   playerRegistrationSchema("CreateGameRequest"),
		"JoinGameRequest":     playerRegistrationSchema("JoinGameRequest"),
		"MoveRequest":         moveRequestSchema(),
		"PlayerActionRequest": playerActionRequestSchema(),
		"GameSummary":         gameSummarySchema(),
		"ListGamesResponse":   listGamesResponseSchema(),
		"GameState":           gameStateSchema(),
		"PlayerResponse":      playerResponseSchema(),
		"SessionResponse":     sessionResponseSchema(),
		"ErrorResponse":       errorResponseSchema(),
		"MoveResponse":        moveResponseSchema(),
		"MoveResponseDTO":     moveResponseDTOSchema(),
		"Position":            positionSchema(),
		"Piece":               pieceSchema(),
		"Result":              resultSchema(),
		"GameResponse":        gameResponseSchema(),
		"PlayerGameResponse":  playerGameResponseSchema(),
		"MoveHistoryResponse": moveHistoryResponseSchema(),
		"ValidMovesResponse":  validMovesResponseSchema(),
	}
}

func playerRegistrationSchema(title string) map[string]any {
	return object(title, []string{"player_name", "player_type"}, map[string]any{
		"player_name": map[string]any{"type": "string", "minLength": 1},
		"player_type": map[string]any{"type": "string", "enum": []string{"human", "ai"}},
	})
}

func moveRequestSchema() map[string]any {
	return object("MoveRequest", []string{"player_id", "from", "to"}, map[string]any{
		"player_id": map[string]any{"type": "string"},
		"from":      ref("Position"),
		"to":        ref("Position"),
	})
}

func playerActionRequestSchema() map[string]any {
	return object("PlayerActionRequest", []string{"player_id"}, map[string]any{
		"player_id": map[string]any{"type": "string"},
	})
}

func gameSummarySchema() map[string]any {
	return object("GameSummary", []string{"game_id", "status", "current_turn", "red_player", "black_player", "created_at"}, map[string]any{
		"game_id":      map[string]any{"type": "string"},
		"status":       map[string]any{"type": "string", "enum": []string{"waiting", "active", "completed", "draw"}},
		"current_turn": map[string]any{"type": "string", "enum": []string{"red", "black"}},
		"red_player":   nullableRef("PlayerResponse"),
		"black_player": nullableRef("PlayerResponse"),
		"created_at":   dateTimeSchema(),
	})
}

func listGamesResponseSchema() map[string]any {
	return object("ListGamesResponse", []string{"success", "games"}, map[string]any{
		"success": map[string]any{"type": "boolean"},
		"games":   map[string]any{"type": "array", "items": ref("GameSummary")},
	})
}

func gameStateSchema() map[string]any {
	return object("GameState", []string{"id", "board", "current_turn", "status", "created_at", "updated_at"}, map[string]any{
		"id":           map[string]any{"type": "string"},
		"board":        boardSchema(),
		"red_player":   nullableRef("PlayerResponse"),
		"black_player": nullableRef("PlayerResponse"),
		"current_turn": map[string]any{"type": "string", "enum": []string{"red", "black"}},
		"status":       map[string]any{"type": "string", "enum": []string{"waiting", "active", "completed", "draw"}},
		"result":       nullableRef("Result"),
		"created_at":   dateTimeSchema(),
		"updated_at":   dateTimeSchema(),
	})
}

func boardSchema() map[string]any {
	return map[string]any{
		"type":     "array",
		"minItems": 8,
		"maxItems": 8,
		"items": map[string]any{
			"type":     "array",
			"minItems": 8,
			"maxItems": 8,
			"items": map[string]any{
				"oneOf": []map[string]any{
					ref("Piece"),
					{"type": "null"},
				},
			},
		},
	}
}

func playerResponseSchema() map[string]any {
	return object("PlayerResponse", []string{"id", "name", "color", "type"}, map[string]any{
		"id":    map[string]any{"type": "string"},
		"name":  map[string]any{"type": "string"},
		"color": map[string]any{"type": "string", "enum": []string{"red", "black"}},
		"type":  map[string]any{"type": "string", "enum": []string{"human", "ai"}},
	})
}

func sessionResponseSchema() map[string]any {
	return object("SessionResponse", []string{"token", "expires_at"}, map[string]any{
		"token":      map[string]any{"type": "string"},
		"expires_at": dateTimeSchema(),
	})
}

func errorResponseSchema() map[string]any {
	return object("ErrorResponse", []string{"error", "status_code"}, map[string]any{
		"error":       map[string]any{"type": "string"},
		"status_code": map[string]any{"type": "integer"},
	})
}

func moveResponseSchema() map[string]any {
	return object("MoveResponse", []string{"success", "move", "game_state"}, map[string]any{
		"success":    map[string]any{"type": "boolean"},
		"move":       nullableRef("MoveResponseDTO"),
		"game_state": nullableRef("GameState"),
	})
}

func moveResponseDTOSchema() map[string]any {
	return object("MoveResponseDTO", []string{"from", "to", "player_id", "timestamp", "promoted"}, map[string]any{
		"from":      ref("Position"),
		"to":        ref("Position"),
		"player_id": map[string]any{"type": "string"},
		"timestamp": dateTimeSchema(),
		"captured": map[string]any{
			"type":  []string{"array", "null"},
			"items": ref("Position"),
		},
		"promoted": map[string]any{"type": "boolean"},
	})
}

func positionSchema() map[string]any {
	return object("Position", []string{"row", "col"}, map[string]any{
		"row": map[string]any{"type": "integer", "minimum": 0, "maximum": 7},
		"col": map[string]any{"type": "integer", "minimum": 0, "maximum": 7},
	})
}

func pieceSchema() map[string]any {
	return object("Piece", []string{"id", "color", "is_king"}, map[string]any{
		"id":      map[string]any{"type": "string"},
		"color":   map[string]any{"type": "string", "enum": []string{"red", "black"}},
		"is_king": map[string]any{"type": "boolean"},
	})
}

func resultSchema() map[string]any {
	return object("Result", []string{"winner", "reason"}, map[string]any{
		"winner":     map[string]any{"type": "string"},
		"reason":     map[string]any{"type": "string"},
		"draw_offer": map[string]any{"type": "string"},
	})
}

func gameResponseSchema() map[string]any {
	return object("GameResponse", []string{"success", "game_id", "game_state"}, map[string]any{
		"success":    map[string]any{"type": "boolean"},
		"game_id":    map[string]any{"type": "string"},
		"game_state": nullableRef("GameState"),
	})
}

func playerGameResponseSchema() map[string]any {
	return object("PlayerGameResponse", []string{"success", "game_id", "player", "game_state"}, map[string]any{
		"success":    map[string]any{"type": "boolean"},
		"game_id":    map[string]any{"type": "string"},
		"player":     nullableRef("PlayerResponse"),
		"session":    nullableRef("SessionResponse"),
		"game_state": nullableRef("GameState"),
	})
}

func moveHistoryResponseSchema() map[string]any {
	return object("MoveHistoryResponse", []string{"success", "moves"}, map[string]any{
		"success": map[string]any{"type": "boolean"},
		"moves":   map[string]any{"type": "array", "items": ref("MoveResponseDTO")},
	})
}

func validMovesResponseSchema() map[string]any {
	return object("ValidMovesResponse", []string{"success", "moves"}, map[string]any{
		"success": map[string]any{"type": "boolean"},
		"moves":   map[string]any{"type": "array", "items": ref("Position")},
	})
}

func object(title string, required []string, properties map[string]any) map[string]any {
	return map[string]any{
		"title":                title,
		"type":                 "object",
		"additionalProperties": false,
		"required":             required,
		"properties":           properties,
	}
}

func nullableRef(schemaName string) map[string]any {
	return map[string]any{
		"oneOf": []map[string]any{
			ref(schemaName),
			{"type": "null"},
		},
	}
}

func dateTimeSchema() map[string]any {
	return map[string]any{"type": "string", "format": "date-time"}
}
