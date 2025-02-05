package system

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// Redis kapselt den go-redis Client und den Context.
type Redis struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedis liest Umgebungsvariablen aus (z. B. REDIS_ADDR, REDIS_PASSWORD, REDIS_DB)
// und stellt eine Verbindung zu Redis her.
func NewRedis() *Redis {
	// Standardwerte, falls keine Env-Variablen gesetzt sind.
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if n, err := strconv.Atoi(dbStr); err == nil {
			db = n
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	return &Redis{
		client: client,
		ctx:    ctx,
	}
}

// Set serialisiert den Wert (value) als JSON und speichert ihn unter dem angegebenen Key.
func (r *Redis) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, 0).Err() // 0 = kein Ablaufdatum
}

// Get lädt den Wert für den gegebenen Key, deserialisiert ihn aus JSON
// und schreibt das Ergebnis in target (das als Pointer übergeben werden muss).
func (r *Redis) Get(key string, target interface{}) error {
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), target)
}

// LPush fügt einen Wert am Anfang der Liste hinzu.
func (r *Redis) LPush(key string, value interface{}) error {
	return r.client.LPush(r.ctx, key, value).Err()
}

// LRange gibt die Elemente der Liste zwischen start und stop zurück.
func (r *Redis) LRange(key string, start, stop int64) ([]string, error) {
	return r.client.LRange(r.ctx, key, start, stop).Result()
}

// Del löscht den angegebenen Key.
func (r *Redis) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// LRem entfernt Elemente aus der Liste, die dem übergebenen Wert entsprechen.
// count steuert dabei, wie viele Vorkommen entfernt werden sollen.
func (r *Redis) LRem(key string, count int64, value interface{}) error {
	return r.client.LRem(r.ctx, key, count, value).Err()
}
