# API Verification Report

Date: 2026-07-08

## Summary

API verification could not be executed because the project could not be built in this environment. All requested BDR API checks are therefore marked **FAILED** due to the server not being runnable. No API responses are fabricated in this report.

## Environment and Build Evidence

### STEP 1: Go Installation Check

Command:

```sh
go version
```

Result:

```text
/bin/bash: line 1: go: command not found
```

Additional local toolchain check:

```sh
PATH=/tmp/go/bin:$PATH go version
```

Result:

```text
Segmentation fault
```

The repository requires Go 1.24:

```text
go 1.24
```

Attempted installation routes:

```sh
sudo -n apt-get update
```

Result:

```text
sudo: The "no new privileges" flag is set, which prevents sudo from running as root.
sudo: If sudo is running in a container, you may need to adjust the container configuration to disable the flag.
```

```sh
apt-get download golang-1.24-go golang-1.24-src
```

Result:

```text
Temporary failure resolving 'us.archive.ubuntu.com'
E: Failed to fetch http://us.archive.ubuntu.com/ubuntu/pool/universe/g/golang-1.24/golang-1.24-go_1.24.13-2_amd64.deb  Temporary failure resolving 'us.archive.ubuntu.com'
E: Failed to fetch http://us.archive.ubuntu.com/ubuntu/pool/universe/g/golang-1.24/golang-1.24-src_1.24.13-2_all.deb  Temporary failure resolving 'us.archive.ubuntu.com'
```

Docker fallback:

```sh
docker images --format '{{.Repository}}:{{.Tag}} {{.ID}}'
```

Result:

```text
permission denied while trying to connect to the docker API at unix:///var/run/docker.sock
```

### STEP 2: Build

Command:

```sh
PATH=/tmp/go/bin:$PATH go mod tidy && PATH=/tmp/go/bin:$PATH go build -o agent-checkers ./src
```

Result:

```text
/bin/bash: line 1:     3 Segmentation fault         PATH=/tmp/go/bin:$PATH go mod tidy
```

### STEP 3: Server Startup

Result: **FAILED**

Evidence:

```text
agent-checkers: cannot open `agent-checkers' (No such file or directory)
```

Because the binary could not be built and no prebuilt `agent-checkers` binary existed, the server could not be started on port 8080.

## BDR Results

## BDR-001: Player Registration

Status: **FAILED**

Reason: API calls could not be executed because the server could not be built or started.

Requested API calls:

```sh
curl -X POST http://localhost:8080/api/v1/games \
  -H 'Content-Type: application/json' \
  -d '{"player_name":"Alice","player_type":"human"}'

curl -X POST http://localhost:8080/api/v1/games/{game_id}/join \
  -H 'Content-Type: application/json' \
  -d '{"player_name":"Bob","player_type":"human"}'
```

Actual API response: not available; server was not running.

Static implementation note: `POST /api/v1/games` and `POST /api/v1/games/{id}/join` are registered in `src/api/handlers/handlers.go`.

## BDR-002: Move Validation

Status: **FAILED**

Reason: API calls could not be executed because the server could not be built or started.

Requested API calls:

```sh
curl http://localhost:8080/api/v1/games/{game_id}

curl -X POST http://localhost:8080/api/v1/games/{game_id}/moves \
  -H 'Content-Type: application/json' \
  -d '{"from":{"row":2,"col":3},"to":{"row":3,"col":4}}'
```

Actual API response: not available; server was not running.

Static implementation notes:

- `POST /api/v1/games/{id}/moves` is registered.
- `MoveRequest` requires `player_id`; the requested example omits it, so that exact payload would not identify the moving player.
- Invalid move errors are returned as JSON error payloads by the handler.

## BDR-003: Game State Management

Status: **FAILED**

Reason: API calls could not be executed because the server could not be built or started.

Requested API call:

```sh
curl http://localhost:8080/api/v1/games/{game_id}
```

Actual API response: not available; server was not running.

Static implementation notes:

- `GET /api/v1/games/{id}` is registered.
- `GameState` includes `board`, `current_turn`, players, status, result, and timestamps.
- `GameState` does not include `move_history`; move history is exposed separately at `GET /api/v1/games/{id}/moves`.

## BDR-004: Win/Lose/Draw

Status: **FAILED**

Reason: API calls could not be executed because the server could not be built or started.

Requested API calls:

```sh
curl -X DELETE http://localhost:8080/api/v1/games/{game_id}
curl -X POST http://localhost:8080/api/v1/games/{game_id}/draw
```

Actual API response: not available; server was not running.

Static implementation notes:

- `DELETE /api/v1/games/{id}` is registered for resignation, but it expects a JSON body containing `player_id`.
- `POST /api/v1/games/{id}/draw` is registered and expects a JSON body containing `player_id`.
- Terminal win status is represented as `completed` with a `result.winner`, not as `red_wins` or `black_wins`.

## BDR-009: Concurrent Games

Status: **FAILED**

Reason: API calls could not be executed because the server could not be built or started.

Requested checks:

- Create multiple games.
- Verify isolation between games.
- List games by status.

Actual API response: not available; server was not running.

Static implementation notes:

- The in-memory store supports multiple games and `ListGames`.
- No REST route for `GET /api/v1/games?status=waiting` is registered in `src/api/handlers/handlers.go`, so the requested list-by-status API is not currently exposed.

## Conclusion

The verification run is blocked at build time by the absence of a usable Go 1.24+ toolchain. The requested API behavior tests should be rerun after a working Go toolchain is available or a prebuilt `agent-checkers` binary is supplied.
