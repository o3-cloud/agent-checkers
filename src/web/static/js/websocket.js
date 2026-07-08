/**
 * websocket.js — thin WebSocket wrapper with reconnect and ping/pong.
 */
(function (global) {
  "use strict";

  function WSClient(url) {
    this.url = url;
    this.ws = null;
    this.handlers = {};
    this.reconnectDelay = 2000;
    this.shouldReconnect = true;
    this.pingInterval = null;
  }

  WSClient.prototype.connect = function () {
    var self = this;
    this.shouldReconnect = true;

    try {
      this.ws = new WebSocket(this.url);
    } catch (e) {
      this._scheduleReconnect();
      return;
    }

    this.ws.onopen = function () {
      self._emit("open", null);
      self._startPing();
    };

    this.ws.onmessage = function (event) {
      var data;
      try {
        data = JSON.parse(event.data);
      } catch (e) {
        return;
      }
      if (data.type === "ping") {
        self._sendPong();
        return;
      }
      self._emit(data.type, data.payload);
    };

    this.ws.onclose = function () {
      self._emit("close", null);
      self._stopPing();
      if (self.shouldReconnect) {
        self._scheduleReconnect();
      }
    };

    this.ws.onerror = function () {
      self._emit("error", null);
    };
  };

  WSClient.prototype.on = function (type, cb) {
    if (!this.handlers[type]) {
      this.handlers[type] = [];
    }
    this.handlers[type].push(cb);
  };

  WSClient.prototype.send = function (data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  };

  WSClient.prototype.close = function () {
    this.shouldReconnect = false;
    this._stopPing();
    if (this.ws) {
      this.ws.close();
    }
  };

  // --- internals ---

  WSClient.prototype._emit = function (type, payload) {
    var cbs = this.handlers[type];
    if (!cbs) return;
    for (var i = 0; i < cbs.length; i++) {
      try {
        cbs[i](payload);
      } catch (e) {
        // swallow handler errors
      }
    }
  };

  WSClient.prototype._scheduleReconnect = function () {
    var self = this;
    setTimeout(function () {
      if (self.shouldReconnect) {
        self.connect();
      }
    }, this.reconnectDelay);
  };

  WSClient.prototype._startPing = function () {
    var self = this;
    this._stopPing();
    this.pingInterval = setInterval(function () {
      self._sendPong(); // acts as keepalive; server also pings
    }, 25000);
  };

  WSClient.prototype._stopPing = function () {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  };

  WSClient.prototype._sendPong = function () {
    this.send({ type: "pong", payload: { timestamp: new Date().toISOString() } });
  };

  global.WSClient = WSClient;
})(window);
