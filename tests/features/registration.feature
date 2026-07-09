Feature: Game Registration
  As a player (human or AI agent)
  I want to register for a checkers game
  So that I can play against another player

  Background:
    Given the game server is running
    And the game state is empty

  Scenario: First player registers and game is waiting
    When I register a player with name "Alice" and type "human"
    Then the response status should be 201
    And the response should contain a game_id
    And the response should contain a player_id
    And the player should be assigned color "red"
    And the game status should be "waiting"
    And the game state should have current_turn "red"

  Scenario: Second player joins and game starts
    Given a game is waiting for an opponent with player "Alice"
    When player "Claude" joins the game with type "ai"
    Then the response status should be 200
    And the response should contain a game_id
    And the response should contain a player_id
    And the player should be assigned color "black"
    And the game status should be "active"
    And the game state should have current_turn "red"

  Scenario: AI vs AI game registration
    Given a game is waiting for an opponent with player "Agent1" of type "ai"
    When player "Agent2" joins the game with type "ai"
    Then the response status should be 200
    And the game status should be "active"
    And the red player should be "Agent1"
    And the black player should be "Agent2"

  Scenario: Registration with empty name fails
    When I register a player with name "" and type "human"
    Then the response status should be 400
    And the response should contain an error

  Scenario: Joining a non-existent game fails
    When player "Bob" joins game "nonexistent-game-id" with type "human"
    Then the response status should be 404
    And the response should contain an error
