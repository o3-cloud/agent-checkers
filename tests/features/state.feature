Feature: Game State Management
  As a player or game observer
  I want to query the current game state
  So that I can understand the board position and whose turn it is

  Background:
    Given the game server is running
    And a game exists with two players

  Scenario: Query game state via API
    When I request the game state
    Then the response status should be 200
    And the response should contain a game_id
    And the game state should have current_turn "red"
    And the game state should have status "active"
    And the board should be an 8x8 grid
    And red pieces should occupy rows 0-2
    And black pieces should occupy rows 5-7
    And each piece should have id, color, and is_king fields

  Scenario: Query non-existent game returns 404
    When I request the game state for game "nonexistent-id"
    Then the response status should be 404
    And the response should contain an error

  Scenario: Board state representation
    When I request the game state
    Then the board should be an 8x8 grid
    And the board should have 12 red pieces
    And the board should have 12 black pieces
    And empty squares should be null
    And all pieces should have is_king false at start

  Scenario: Health check
    When I request the health endpoint
    Then the response status should be 200
    And the response should contain status "ok"
