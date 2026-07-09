Feature: WebSocket Real-Time Game Updates
  As a player
  I want to receive real-time updates when game state changes
  So that I can see moves and events as they happen

  Background:
    Given the game server is running
    And a game exists with two players

  Scenario: WebSocket connection sends current game state
    When I connect a WebSocket to the game
    Then I should receive a "game_state" event
    And the event should contain the game_id

  Scenario: WebSocket rejects invalid session token
    When I connect a WebSocket with an invalid token
    Then the WebSocket connection should be rejected with status 401

  Scenario: Move broadcast via WebSocket
    Given a red piece is at position row 2 col 3
    And I have connected a WebSocket to the game
    When the red player moves from row 2 col 3 to row 3 col 4
    Then I should receive a "move_made" event
    And I should receive a "turn_changed" event
