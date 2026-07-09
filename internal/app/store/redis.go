// Package store provides persistence contracts and Redis storage.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stackable-specs/agent-checkers/internal/app/game"
)

const defaultRedisPrefix = "agent-checkers"

// RedisStore stores games, players, and move history in Redis.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedisStore creates a Redis-backed store and verifies connectivity.
func NewRedisStore(addr, password string, db, poolSize int, prefix string) (*RedisStore, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		addr = "localhost:6379"
	}
	if poolSize <= 0 {
		poolSize = 10
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = defaultRedisPrefix
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})

	ctx, cancel := redisContext()
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w; close redis client: %v", err, closeErr)
		}
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisStore{client: client, prefix: prefix}, nil
}

// SaveGame persists a game state.
func (r *RedisStore) SaveGame(g *game.Game) error {
	if g == nil {
		return errors.New("game is nil")
	}
	if g.ID == "" {
		return errors.New("game ID is empty")
	}

	data, err := json.Marshal(g.Clone())
	if err != nil {
		return fmt.Errorf("marshal game %q: %w", g.ID, err)
	}

	ctx, cancel := redisContext()
	defer cancel()

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.gameKey(g.ID), data, 0)
	pipe.Del(ctx, r.movesKey(g.ID))
	for _, move := range g.Moves {
		moveData, err := json.Marshal(cloneMove(move))
		if err != nil {
			return fmt.Errorf("marshal move for game %q: %w", g.ID, err)
		}
		pipe.RPush(ctx, r.movesKey(g.ID), moveData)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("save game %q: %w", g.ID, err)
	}
	return nil
}

// LoadGame retrieves a game by ID.
func (r *RedisStore) LoadGame(id string) (*game.Game, error) {
	ctx, cancel := redisContext()
	defer cancel()

	data, err := r.client.Get(ctx, r.gameKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("load game %q: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("load game %q: %w", id, err)
	}

	var g game.Game
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("unmarshal game %q: %w", id, err)
	}
	moves, err := r.loadMoveHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load game moves %q: %w", id, err)
	}
	g.Moves = moves
	return g.Clone(), nil
}

// DeleteGame removes a game by ID.
func (r *RedisStore) DeleteGame(id string) error {
	ctx, cancel := redisContext()
	defer cancel()

	deleted, err := r.client.Del(ctx, r.gameKey(id), r.movesKey(id)).Result()
	if err != nil {
		return fmt.Errorf("delete game %q: %w", id, err)
	}
	if deleted == 0 {
		return fmt.Errorf("delete game %q: %w", id, ErrNotFound)
	}
	return nil
}

// ListGames returns games matching the provided filter.
func (r *RedisStore) ListGames(filter GameFilter) ([]*game.Game, error) {
	ctx, cancel := redisContext()
	defer cancel()

	iter := r.client.Scan(ctx, 0, r.gameKey("*"), 0).Iterator()
	games := make([]*game.Game, 0)
	for iter.Next(ctx) {
		key := iter.Val()
		if strings.HasSuffix(key, ":moves") {
			continue
		}
		data, err := r.client.Get(ctx, key).Bytes()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("load listed game %q: %w", key, err)
		}

		var g game.Game
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, fmt.Errorf("unmarshal listed game %q: %w", key, err)
		}
		moves, err := r.loadMoveHistory(ctx, g.ID)
		if err != nil {
			return nil, fmt.Errorf("load listed game moves %q: %w", g.ID, err)
		}
		g.Moves = moves
		if (filter.StatusSet || filter.Status != 0) && g.Status != filter.Status {
			continue
		}
		if filter.PlayerID != "" && !gameHasPlayer(&g, filter.PlayerID) {
			continue
		}
		games = append(games, g.Clone())
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan games: %w", err)
	}
	return games, nil
}

// SavePlayer persists a player.
func (r *RedisStore) SavePlayer(p *game.Player) error {
	if p == nil {
		return errors.New("player is nil")
	}
	if p.ID == "" {
		return errors.New("player ID is empty")
	}

	data, err := json.Marshal(clonePlayer(p))
	if err != nil {
		return fmt.Errorf("marshal player %q: %w", p.ID, err)
	}

	ctx, cancel := redisContext()
	defer cancel()
	if err := r.client.Set(ctx, r.playerKey(p.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("save player %q: %w", p.ID, err)
	}
	return nil
}

// LoadPlayer retrieves a player by ID.
func (r *RedisStore) LoadPlayer(id string) (*game.Player, error) {
	ctx, cancel := redisContext()
	defer cancel()

	data, err := r.client.Get(ctx, r.playerKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("load player %q: %w", id, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("load player %q: %w", id, err)
	}

	var p game.Player
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal player %q: %w", id, err)
	}
	return clonePlayer(&p), nil
}

// AppendMove appends a move to a stored game's history.
func (r *RedisStore) AppendMove(gameID string, move game.Move) error {
	ctx, cancel := redisContext()
	defer cancel()

	exists, err := r.client.Exists(ctx, r.gameKey(gameID)).Result()
	if err != nil {
		return fmt.Errorf("append move for game %q: %w", gameID, err)
	}
	if exists == 0 {
		return fmt.Errorf("append move for game %q: %w", gameID, ErrNotFound)
	}

	moveData, err := json.Marshal(cloneMove(move))
	if err != nil {
		return fmt.Errorf("marshal move for game %q: %w", gameID, err)
	}
	if err := r.client.RPush(ctx, r.movesKey(gameID), moveData).Err(); err != nil {
		return fmt.Errorf("append move for game %q: %w", gameID, err)
	}
	return nil
}

// GetMoveHistory retrieves the move history for a game.
func (r *RedisStore) GetMoveHistory(gameID string) ([]game.Move, error) {
	ctx, cancel := redisContext()
	defer cancel()

	exists, err := r.client.Exists(ctx, r.gameKey(gameID)).Result()
	if err != nil {
		return nil, fmt.Errorf("get move history for game %q: %w", gameID, err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("get move history for game %q: %w", gameID, ErrNotFound)
	}

	moves, err := r.loadMoveHistory(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("get move history for game %q: %w", gameID, err)
	}
	return moves, nil
}

func (r *RedisStore) loadMoveHistory(ctx context.Context, gameID string) ([]game.Move, error) {
	values, err := r.client.LRange(ctx, r.movesKey(gameID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	moves := make([]game.Move, 0, len(values))
	for _, value := range values {
		var move game.Move
		if err := json.Unmarshal([]byte(value), &move); err != nil {
			return nil, fmt.Errorf("unmarshal move: %w", err)
		}
		moves = append(moves, cloneMove(move))
	}
	return moves, nil
}

func (r *RedisStore) gameKey(id string) string {
	return r.prefix + ":games:" + id
}

func (r *RedisStore) playerKey(id string) string {
	return r.prefix + ":players:" + id
}

func (r *RedisStore) movesKey(id string) string {
	return r.prefix + ":games:" + id + ":moves"
}

func redisContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
