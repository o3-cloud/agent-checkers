Feature: MCP Server Interface
  As an AI agent
  I want to interact with the game via MCP tools
  So that I can make moves programmatically

  Background:
    Given the MCP server is available
    And the game state is empty

  Scenario: MCP tool - register_player
    When the agent calls "register_player" with name "Claude" and type "ai"
    Then the MCP response should contain a player_id
    And the MCP response should contain a game_id
    And the MCP response should have color "red"
    And the MCP response should have success true

  Scenario: MCP tool - get_game_state
    Given a game exists with two players via MCP
    When the agent calls "get_game_state" with the game_id
    Then the MCP response should contain a game_state
    And the game state should have current_turn "red"
    And the game state should have status "active"

  Scenario: MCP tool - get_valid_moves
    Given a game exists with two players via MCP
    When the agent calls "get_valid_moves" with the game_id
    Then the MCP response should contain a list of moves

  Scenario: MCP tool - tools/list
    When the agent calls "tools/list"
    Then the MCP response should contain a list of tools
    And the tools should include "register_player"
    And the tools should include "get_game_state"
    And the tools should include "make_move"
    And the tools should include "get_valid_moves"
    And the tools should include "resign"
    And the tools should include "offer_draw"
    And the tools should include "accept_draw"

  Scenario: MCP tool - make_move
    Given a game exists with two players via MCP
    And a red piece is at position row 2 col 3
    When the agent calls "make_move" with from row 2 col 3 to row 3 col 4
    Then the MCP response should have success true
    And the MCP response should contain a game_state
    And the game state should have current_turn "black"

  Scenario: MCP tool - resign
    Given a game exists with two players via MCP
    When the agent calls "resign" with the red player's id
    Then the MCP response should have success true
    And the game status should be "completed"
    And the winner should be "black"

  Scenario: MCP tool - unknown tool returns error
    When the agent calls "unknown_tool"
    Then the MCP response should contain an error
