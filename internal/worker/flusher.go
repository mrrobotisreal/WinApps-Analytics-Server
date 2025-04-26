package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type rawEvent map[string]any

func StartFlusher(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			flush(ctx, db, rdb)
		case <-ctx.Done():
			return
		}
	}
}

func flush(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client) {
	for {
		data, err := rdb.LPop(ctx, "event_queue").Result()
		if err == redis.Nil {
			return
		}
		if err != nil {
			log.Printf("redis pop error: %v", err)
			return
		}
		var ev rawEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			log.Printf("json parse err: %v", err)
			continue
		}

		_, err = db.Exec(ctx, `INSERT INTO analytics_events (api_key, event_type, url, referrer, lang, screen_w, screen_h, session_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			ev["api_key"], ev["event_type"], ev["url"], ev["referrer"], ev["lang"], ev["screen_w"], ev["screen_h"], ev["session_id"],
		)
		if err != nil {
			log.Printf("pg insert err: %v", err)
		}
	}
}
