/**
 * app.js — main application logic: REST calls, game state, UI orchestration.
 */
(function (global) {
  "use strict";

  var App = {};

  // --- State ---
  App.gameID = null;
  App.playerID = null;
  App.sessionToken = null;
  App.playerColor = null;
  App.gameState = null;
  App.ws = null;
  App.sessionStorageKey = "agent-checkers-session";

  // --- Session persistence ---

  App._isEndedStatus = function (status) {
    return status === "completed" || status === "draw" || status === "resigned";
  };

  App._isGameEnded = function () {
    return App.gameState && App._isEndedStatus(App.gameState.status);
  };

  App._setStatusText = function (text) {
    var bar = document.getElementById("status-bar");
    if (bar) {
      bar.textContent = text;
    }
  };

  App.saveSession = function () {
    if (!App.gameID || !App.playerID || !App.sessionToken || !App.playerColor) return;

    try {
      localStorage.setItem(
        App.sessionStorageKey,
        JSON.stringify({
          gameID: App.gameID,
          playerID: App.playerID,
          sessionToken: App.sessionToken,
          playerColor: App.playerColor,
        })
      );
    } catch (e) {
      // localStorage may be unavailable in private browsing or restricted contexts.
    }
  };

  App.loadSession = function () {
    var raw;
    var session;

    try {
      raw = localStorage.getItem(App.sessionStorageKey);
      if (!raw) return false;
      session = JSON.parse(raw);
    } catch (e) {
      App.clearSession();
      return false;
    }

    if (!session || !session.gameID || !session.playerID || !session.sessionToken || !session.playerColor) {
      App.clearSession();
      return false;
    }

    App.gameID = session.gameID;
    App.playerID = session.playerID;
    App.sessionToken = session.sessionToken;
    App.playerColor = session.playerColor;
    return true;
  };

  App.clearSession = function () {
    try {
      localStorage.removeItem(App.sessionStorageKey);
    } catch (e) {
      // localStorage may be unavailable in private browsing or restricted contexts.
    }
  };

  App.hasStoredSession = function () {
    var raw;
    var session;

    try {
      raw = localStorage.getItem(App.sessionStorageKey);
      if (!raw) return false;
      session = JSON.parse(raw);
    } catch (e) {
      App.clearSession();
      return false;
    }

    if (!session || !session.gameID || !session.playerID || !session.sessionToken || !session.playerColor) {
      App.clearSession();
      return false;
    }
    return true;
  };

  App.init = function () {
    BoardUI.onSquareClick = App.selectSquare;

    // Close resign modal on Escape key
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" || e.keyCode === 27) {
        var resignModal = document.getElementById("resign-modal");
        if (resignModal && resignModal.style.display !== "none") {
          App.hideResignModal();
        }
        var joinForm = document.getElementById("join-form");
        if (joinForm && joinForm.style.display !== "none") {
          App.hideJoinForm();
        }
      }
    });

    // Clear join form error when user types
    var joinGameIdInput = document.getElementById("join-game-id");
    var joinPlayerNameInput = document.getElementById("join-player-name");
    var joinErrorEl = document.getElementById("join-error");
    if (joinGameIdInput) {
      joinGameIdInput.addEventListener("input", function () {
        if (joinErrorEl) joinErrorEl.textContent = "";
      });
    }
    if (joinPlayerNameInput) {
      joinPlayerNameInput.addEventListener("input", function () {
        if (joinErrorEl) joinErrorEl.textContent = "";
      });
    }

    if (!App.hasStoredSession()) {
      App._setStatusText("No game loaded");
      BoardUI.render(null);
      return;
    }

    App._setStatusText("Reconnecting...");
    if (!App.loadSession()) {
      App._setStatusText("No game loaded");
      BoardUI.render(null);
      return;
    }

    App.loadGame(App.gameID, true).then(function () {
      if (App.gameState && !App._isGameEnded()) {
        App.connectWebSocket();
      }
    }).catch(function () {
      App.clearSession();
      App.gameID = null;
      App.gameState = null;
      App.refreshUI();
    });
  };

  // --- API helpers ---

  App._api = function (method, path, body) {
    return fetch(path, {
      method: method,
      headers: { "Content-Type": "application/json" },
      body: body ? JSON.stringify(body) : undefined,
    }).then(function (resp) {
      return resp.json().then(function (data) {
        data._status = resp.status;
        return data;
      });
    });
  };

  // --- New Game ---

  App.showNewGameForm = function () {
    var errorEl = document.getElementById("new-game-error");
    if (errorEl) errorEl.textContent = "";
    document.getElementById("new-game-form").style.display = "flex";
  };

  App.hideNewGameForm = function () {
    var errorEl = document.getElementById("new-game-error");
    if (errorEl) errorEl.textContent = "";
    document.getElementById("new-game-form").style.display = "none";
  };

  App.createGame = function () {
    var name = document.getElementById("new-game-name").value || "Player1";
    var type = document.getElementById("new-game-type").value;
    var errorEl = document.getElementById("new-game-error");
    if (!name.trim()) {
      if (errorEl) errorEl.textContent = "Player name is required";
      return;
    }
    App._api("POST", "/api/v1/games", {
      player_name: name,
      player_type: type,
    }).then(function (data) {
      if (data._status !== 201) {
        if (errorEl) errorEl.textContent = data.error || "Failed to create game";
        return;
      }
      if (errorEl) errorEl.textContent = "";
      App.gameID = data.game_id;
      App.playerID = data.player.id;
      App.playerColor = data.player.color;
      if (data.session) {
        App.sessionToken = data.session.token;
      }
      App.gameState = data.game_state;
      App.saveSession();
      App.hideNewGameForm();
      App.refreshUI();
      App.connectWebSocket();
    }).catch(function (err) {
      if (errorEl) errorEl.textContent = "Unable to connect to server";
    });
  };

  // --- Join Game ---

  App.showJoinForm = function () {
    var errorEl = document.getElementById("join-error");
    if (errorEl) errorEl.textContent = "";
    document.getElementById("join-form").style.display = "flex";
    App.fetchWaitingGames();
  };

  // --- Waiting Games List (BDR-017) ---

  App.fetchWaitingGames = function () {
    var listEl = document.getElementById("waiting-games-list");
    if (!listEl) return;
    listEl.innerHTML = "<p class="loading-text">Loading...</p>";
    App._api("GET", "/api/v1/games?status=waiting", null).then(function (data) {
      if (data._status !== 200) {
        listEl.innerHTML = "<p class="empty-text">Unable to load games</p>";
        return;
      }
      var games = data.games || [];
      if (games.length === 0) {
        listEl.innerHTML = "<p class="empty-text">No games waiting — create a new one!</p>";
        return;
      }
      listEl.innerHTML = "";
      for (var i = 0; i < games.length; i++) {
        var g = games[i];
        var entry = document.createElement("div");
        entry.className = "waiting-game-entry";
        var redName = (g.red_player && g.red_player.name) ? g.red_player.name : "Unknown";
        var gameIdShort = g.game_id ? g.game_id.substring(0, 8) + "..." : "?";
        entry.innerHTML = "<span class="game-id-short">" + gameIdShort + "</span> <span class="red-player">" + redName + "</span>";
        (function (fullId) {
          entry.addEventListener("click", function () {
            var gameInput = document.getElementById("join-game-id");
            if (gameInput) {
              gameInput.value = fullId;
              var nameInput = document.getElementById("join-player-name");
              if (nameInput) nameInput.focus();
            }
          });
        })(g.game_id);
        listEl.appendChild(entry);
      }
    }).catch(function () {
      listEl.innerHTML = "<p class="empty-text">Unable to load games</p>";
    });
  };

  App.hideJoinForm = function () {
    var errorEl = document.getElementById("join-error");
    if (errorEl) errorEl.textContent = "";
    document.getElementById("join-form").style.display = "none";
  };

  App.joinGame = function () {
    var gameID = document.getElementById("join-game-id").value;
    var name = document.getElementById("join-player-name").value || "Player2";
    var type = document.getElementById("join-player-type").value;
    var errorEl = document.getElementById("join-error");
    if (!gameID) {
      if (errorEl) errorEl.textContent = "Game ID is required";
      return;
    }
    App._api("POST", "/api/v1/games/" + gameID + "/join", {
      player_name: name,
      player_type: type,
    }).then(function (data) {
      if (data._status !== 200) {
        if (errorEl) {
          errorEl.textContent = data.error || "Failed to join game";
        }
        return;
      }
      if (errorEl) errorEl.textContent = "";
      App.gameID = data.game_id;
      App.playerID = data.player.id;
      App.playerColor = data.player.color;
      if (data.session) {
        App.sessionToken = data.session.token;
      }
      App.gameState = data.game_state;
      App.saveSession();
      App.hideJoinForm();
      App.refreshUI();
      App.connectWebSocket();
    }).catch(function (err) {
      if (errorEl) {
        errorEl.textContent = "Unable to connect to server";
      }
    });
  };

  // --- Load Game ---

  App.loadGame = function (gameID, suppressAlert) {
    return App._api("GET", "/api/v1/games/" + gameID, null).then(function (data) {
      if (data._status !== 200) {
        App.clearSession();
        App.gameID = null;
        App.gameState = null;
        if (!suppressAlert) {
          alert("Failed to load game: " + (data.error || "unknown"));
        }
        App.refreshUI();
        return;
      }
      App.gameID = data.game_id;
      App.gameState = data.game_state;
      App.refreshUI();
      return data;
    });
  };

  // --- Square selection and valid moves ---

  App.selectSquare = function (row, col) {
    if (!App.gameState || App.gameState.status !== "active") return;

    var pieceData = App.gameState.board[row] && App.gameState.board[row][col];

    // If we have a selection and clicked a valid destination, make the move
    if (BoardUI.selectedRow !== null && BoardUI.selectedCol !== null) {
      var isValidDest = false;
      for (var i = 0; i < BoardUI.validMoves.length; i++) {
        if (BoardUI.validMoves[i].row === row && BoardUI.validMoves[i].col === col) {
          isValidDest = true;
          break;
        }
      }
      if (isValidDest) {
        App.makeMove(
          { row: BoardUI.selectedRow, col: BoardUI.selectedCol },
          { row: row, col: col }
        );
        return;
      }
    }

    // If clicking own piece, select it
    if (pieceData && pieceData.color === App.playerColor) {
      App._fetchValidMoves(row, col);
      return;
    }

    // Otherwise clear selection
    BoardUI.clearSelection();
    BoardUI.render(App.gameState);
  };

  App._fetchValidMoves = function (row, col) {
    var url = "/api/v1/games/" + App.gameID + "/valid-moves?row=" + row + "&col=" + col;
    App._api("GET", url, null).then(function (data) {
      if (data._status !== 200) {
        BoardUI.clearSelection();
        BoardUI.render(App.gameState);
        return;
      }
      BoardUI.setSelection(row, col, data.moves || []);
      BoardUI.render(App.gameState);
    });
  };

  // --- Make move ---

  App.makeMove = function (from, to) {
    App._api("POST", "/api/v1/games/" + App.gameID + "/moves", {
      player_id: App.playerID,
      from: from,
      to: to,
    }).then(function (data) {
      if (data._status !== 200) {
        alert("Invalid move: " + (data.error || "unknown"));
        BoardUI.clearSelection();
        BoardUI.render(App.gameState);
        return;
      }
      App.gameState = data.game_state;
      BoardUI.clearSelection();
      App.refreshUI();
    });
  };

  // --- Resign ---

  App.resignGame = function () {
    if (!App.gameID || !App.playerID) return;
    var modal = document.getElementById("resign-modal");
    if (modal) {
      modal.style.display = "flex";
    }
  };

  App.hideResignModal = function () {
    var modal = document.getElementById("resign-modal");
    if (modal) {
      modal.style.display = "none";
    }
  };

  App.confirmResign = function () {
    if (!App.gameID || !App.playerID) return;
    App.hideResignModal();
    App._api("DELETE", "/api/v1/games/" + App.gameID, {
      player_id: App.playerID,
    }).then(function (data) {
      if (data._status !== 200) {
        alert("Failed to resign: " + (data.error || "unknown"));
        return;
      }
      App.gameState = data.game_state;
      App.clearSession();
      App.refreshUI();
    });
  };

  // --- Status ---

  App.updateStatus = function () {
    var bar = document.getElementById("status-bar");
    if (!App.gameState) {
      bar.textContent = "No game loaded";
      return;
    }

    var status = App.gameState.status;
    var turn = App.gameState.current_turn;
    var result = App.gameState.result;

    if (App._isEndedStatus(status)) {
      App.clearSession();
    }

    if (status === "waiting") {
      bar.textContent = "Waiting for opponent...";
    } else if (status === "active") {
      var whoseTurn = turn === App.playerColor ? "Your turn" : turn + "'s turn";
      bar.textContent = "Active — " + whoseTurn + " (" + turn + ")";
    } else if (status === "completed" && result) {
      bar.textContent = "Winner: " + result.winner + " (" + result.reason + ")";
    } else if (status === "draw") {
      bar.textContent = "Game ended in a draw";
    } else {
      bar.textContent = status;
    }
  };

  // --- Move history ---

  // Convert (row, col) to algebraic notation (e.g., 2,1 -> "b3")
  App._toAlgebraic = function (row, col) {
    var colChar = String.fromCharCode(97 + col); // 'a' = 97
    return colChar + (row + 1);
  };

  // Determine player color for a move based on player ID
  App._getPlayerColor = function (playerId) {
    if (App.gameState) {
      if (App.gameState.red_player && App.gameState.red_player.id === playerId) {
        return "Red";
      }
      if (App.gameState.black_player && App.gameState.black_player.id === playerId) {
        return "Black";
      }
    }
    return "?";
  };

  App.fetchMoveHistory = function () {
    if (!App.gameID) return;
    App._api("GET", "/api/v1/games/" + App.gameID + "/moves", null).then(function (data) {
      if (data._status !== 200) return;
      var list = document.getElementById("move-history");
      if (!list) return;
      list.innerHTML = "";
      var moves = data.moves || [];
      for (var i = 0; i < moves.length; i++) {
        var m = moves[i];
        var li = document.createElement("li");
        var moveNum = i + 1;
        var color = App._getPlayerColor(m.player_id);

        // Build algebraic notation for from/to
        var fromStr = "?";
        var toStr = "?";
        if (m.from && typeof m.from.row === "number" && typeof m.from.col === "number") {
          fromStr = App._toAlgebraic(m.from.row, m.from.col);
        }
        if (m.to && typeof m.to.row === "number" && typeof m.to.col === "number") {
          toStr = App._toAlgebraic(m.to.row, m.to.col);
        }

        // Build move text: "1. Red: b3 → a4"
        var txt = moveNum + ". " + color + ": " + fromStr;

        // Use × for captures, → for regular moves
        var capturedCount = (m.captured && Array.isArray(m.captured)) ? m.captured.length : 0;
        if (capturedCount > 0) {
          txt += " × " + toStr;
          txt += " (capture)";
        } else {
          txt += " → " + toStr;
        }

        // Mark king promotions
        if (m.promoted) {
          txt += " ♚";
        }

        li.textContent = txt;
        list.appendChild(li);
      }

      // Auto-scroll to the latest move
      list.scrollTop = list.scrollHeight;
    });
  };

  // --- Player info ---

  App.renderPlayerInfo = function () {
    var info = document.getElementById("player-info");
    if (!App.gameState) {
      info.innerHTML = "<p>No game loaded</p>";
      return;
    }
    var html = "";
    var red = App.gameState.red_player;
    var black = App.gameState.black_player;
    if (red) {
      html += "<p>Red: " + red.name + " (" + red.type + ")</p>";
    }
    if (black) {
      html += "<p>Black: " + black.name + " (" + black.type + ")</p>";
    }
    if (App.playerColor) {
      html += "<p><strong>You are: " + App.playerColor + "</strong></p>";
    }
    if (App.gameID) {
      html += "<p>Game ID: " + App.gameID.substring(0, 8) + "...</p>";
    }
    info.innerHTML = html;
  };

  // --- Full UI refresh ---

  App.refreshUI = function () {
    App.updateStatus();
    App.renderPlayerInfo();
    BoardUI.render(App.gameState);
    App.fetchMoveHistory();
  };

  // --- WebSocket ---

  App.connectWebSocket = function () {
    if (!App.gameID || !App.sessionToken) return;
    if (App.ws) {
      App.ws.close();
    }
    var proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    var url = proto + "//" + window.location.host + "/api/v1/games/" + App.gameID + "/ws?token=" + App.sessionToken;

    App.ws = new WSClient(url);

    App.ws.on("game_state", function (payload) {
      if (payload && payload.game_state) {
        App.gameState = payload.game_state;
        App.refreshUI();
      }
    });

    App.ws.on("game_started", function (payload) {
      App.loadGame(App.gameID);
    });

    App.ws.on("move_made", function (payload) {
      App.loadGame(App.gameID);
    });

    App.ws.on("turn_changed", function (payload) {
      App.loadGame(App.gameID);
    });

    App.ws.on("game_ended", function (payload) {
      App.clearSession();
      App.loadGame(App.gameID);
    });

    App.ws.connect();
  };

  global.App = App;
})(window);
