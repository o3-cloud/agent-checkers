Feature: Move Validation and Execution
  As a player
  I want to make valid moves and have them rejected if invalid
  So that the game follows standard checkers rules

  Background:
    Given the game server is running
    And a game exists with two players

  Scenario: Valid simple move
    Given a red piece is at position row 2 col 3
    When the red player moves from row 2 col 3 to row 3 col 4
    Then the move should be accepted
    And the response status should be 200
    And the piece should now be at row 3 col 4
    And the turn should pass to "black"

  Scenario: Invalid move - wrong turn
    Given it is red's turn
    When the black player moves from row 5 col 2 to row 4 col 1
    Then the move should be rejected
    And the response status should be 400
    And the response should contain an error

  Scenario: Invalid move - wrong direction
    Given a red piece is at position row 2 col 3
    When the red player moves from row 2 col 3 to row 1 col 4
    Then the move should be rejected
    And the response status should be 400

  Scenario: Valid capture move
    Given a red piece is at position row 2 col 3
    And a black piece is at position row 3 col 4
    And position row 4 col 5 is empty
    When the red player moves from row 2 col 3 to row 4 col 5
    Then the move should be accepted
    And the response status should be 200
    And the black piece at row 3 col 4 should be removed

  Scenario: Mandatory capture - must jump if available
    Given a red piece is at position row 2 col 3
    And a black piece is at position row 3 col 4
    And position row 4 col 5 is empty
    And a red piece is at position row 1 col 2
    And position row 2 col 1 is empty
    When the red player moves from row 1 col 2 to row 2 col 1
    Then the move should be rejected
    And the response status should be 400

  Scenario: King promotion
    Given a red piece is at position row 6 col 3
    And position row 7 col 4 is empty
    When the red player moves from row 6 col 3 to row 7 col 4
    Then the move should be accepted
    And the piece at row 7 col 4 should be a king

  Scenario: Get valid moves for a piece
    When I request valid moves for row 2 col 3
    Then the response status should be 200
    And the response should contain a list of valid moves

  Scenario: Move history is recorded
    Given a red piece is at position row 2 col 3
    When the red player moves from row 2 col 3 to row 3 col 4
    And I request the move history
    Then the response status should be 200
    And the move history should contain 1 move
