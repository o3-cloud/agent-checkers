/**
 * board.js — renders the 8x8 checkers board and handles square clicks.
 */
(function (global) {
  "use strict";

  var BoardUI = {};

  BoardUI.selectedRow = null;
  BoardUI.selectedCol = null;
  BoardUI.validMoves = [];
  BoardUI.onSquareClick = null; // set by app.js

  /**
   * renderBoard(gameState) builds the 8x8 grid.
   * gameState has .board (8x8 array of {color, is_king} or null) and .current_turn.
   */
  BoardUI.render = function (gameState) {
    var container = document.getElementById("board");
    if (!container) return;
    container.innerHTML = "";

    var boardData = gameState ? gameState.board : null;
    var currentTurn = gameState ? gameState.current_turn : null;
    var status = gameState ? gameState.status : null;

    for (var row = 0; row < 8; row++) {
      for (var col = 0; col < 8; col++) {
        var square = document.createElement("div");
        square.className = "square " + ((row + col) % 2 === 1 ? "dark" : "light");
        square.dataset.row = row;
        square.dataset.col = col;

        // Highlight selected square
        if (BoardUI.selectedRow === row && BoardUI.selectedCol === col) {
          square.classList.add("selected");
        }

        // Highlight valid destination squares
        for (var i = 0; i < BoardUI.validMoves.length; i++) {
          if (BoardUI.validMoves[i].row === row && BoardUI.validMoves[i].col === col) {
            square.classList.add("valid-move");
            break;
          }
        }

        // Render piece if present
        if (boardData && boardData[row] && boardData[row][col]) {
          var pieceData = boardData[row][col];
          var piece = document.createElement("div");
          piece.className = "piece " + pieceData.color;
          if (pieceData.is_king) {
            piece.classList.add("king");
          }
          square.appendChild(piece);
        }

        // Only allow interaction on active games
        if (status === "active") {
          (function (r, c) {
            square.addEventListener("click", function () {
              BoardUI._handleClick(r, c);
            });
          })(row, col);
        }

        container.appendChild(square);
      }
    }
  };

  BoardUI._handleClick = function (row, col) {
    if (BoardUI.onSquareClick) {
      BoardUI.onSquareClick(row, col);
    }
  };

  /**
   * setSelection marks a piece as selected and stores valid destination squares.
   */
  BoardUI.setSelection = function (row, col, moves) {
    BoardUI.selectedRow = row;
    BoardUI.selectedCol = col;
    BoardUI.validMoves = moves || [];
  };

  /**
   * clearSelection removes highlight from the currently selected piece.
   */
  BoardUI.clearSelection = function () {
    BoardUI.selectedRow = null;
    BoardUI.selectedCol = null;
    BoardUI.validMoves = [];
  };

  global.BoardUI = BoardUI;
})(window);
