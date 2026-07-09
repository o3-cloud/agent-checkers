Feature: Win/Lose/Draw Conditions
  As a player
  I want the game to correctly detect and announce game endings
  So that there is a clear winner

  Background:
    Given the game server is running
    And a game exists with two players

  Scenario: Win by capturing all opponent pieces
    Given red has 1 piece at row 0 col 1
    And black has 0 pieces remaining
    When I request the game state
    Then the game status should be "completed"
    And the winner should be "red"

  Scenario: Win by blocking all opponent moves
    Given black has pieces but no valid moves
    And red has valid moves available
    When I request the game state
    Then the game status should be "completed"
    And the winner should be "red"

  Scenario: Draw by mutual agreement
    Given the game is in progress
    When player "Alice" offers a draw
    And player "Claude" accepts the draw
    Then the game status should be "draw"

  Scenario: AI agent resignation
    Given the game is in progress
    When the black player resigns
    Then the game status should be "completed"
    And the winner should be "red"
    And the result reason should be "resignation"
