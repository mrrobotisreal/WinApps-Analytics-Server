package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

/* ---------- same Event struct ---------- */
type Event struct {
	APIKey string `json:"apiKey"`
	TS     int64  `json:"ts"`
	URL    string `json:"url"`
	Ref    string `json:"ref"`
	UA     string `json:"ua"`
	Lang   string `json:"lang"`
	SW     int    `json:"sw"`
	SH     int    `json:"sh"`
}

func main() {
	// ── config ──────────────────────────
	brokers := getenv("KAFKA_BROKERS", "kafka:9092")
	topic := getenv("KAFKA_TOPIC", "analytics_events")
	group := getenv("KAFKA_GROUP", "analytics-sink")
	redisURL := getenv("REDIS_URL", "redis://redis:6379/0")
	pgURL := getenv("POSTGRES_URL", "postgres://postgres:postgres@db:5432/analytics?sslmode=disable")
	flushIn := getenvDuration("FLUSH_INTERVAL", 5*time.Second)

	// ── clients ─────────────────────────
	rdb := redis.NewClient(optFromURL(redisURL))
	pool, err := pgxpool.New(context.Background(), pgURL)
	die(err, "postgres")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		GroupID:  group,
		MinBytes: 1e3, // 1 KiB
		MaxBytes: 1e6, // 1 MiB
	})
	defer reader.Close()

	// ── flush ticker ────────────────────
	tick := time.NewTicker(flushIn)
	defer tick.Stop()

	var (
		ctx      = context.Background()
		batch    []Event
		maxBatch = 10_000
	)

	go func() {
		for range tick.C {
			if len(batch) == 0 {
				continue
			}
			if err := flush(ctx, pool, batch); err != nil {
				log.Println("flush error:", err)
				continue
			}
			batch = batch[:0] // reset (keep capacity)
		}
	}()

	log.Printf("consumer group=%s brokers=%s topic=%s", group, brokers, topic)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Fatal("kafka read:", err)
		}

		var ev Event
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Println("malformed json:", err)
			continue
		}
		// ── live counters in Redis (hash per apiKey for today) ──
		key := "dash:" + ev.APIKey + ":" + time.UnixMilli(ev.TS).Format("2006-01-02")
		if err := rdb.HIncrBy(ctx, key, "events", 1).Err(); err != nil {
			log.Println("redis:", err)
		}

		batch = append(batch, ev)
		if len(batch) >= maxBatch {
			if err := flush(ctx, pool, batch); err != nil {
				log.Println("flush error:", err)
			}
			batch = batch[:0]
		}
	}
}

/* ---------- helpers ---------- */

func flush(ctx context.Context, db *pgxpool.Pool, evs []Event) error {
	if len(evs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, e := range evs {
		batch.Queue(`
		  INSERT INTO analytics_events
		  (api_key, event_type, event_time, url, referrer, user_agent,
		   lang, screen_w, screen_h)
		  VALUES ($1,'page_view',to_timestamp($2/1000.0),$3,$4,$5,$6,$7,$8)
		`,
			e.APIKey, e.TS, e.URL, e.Ref, e.UA, e.Lang, e.SW, e.SH)
	}
	br := db.SendBatch(ctx, batch)
	return br.Close()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func die(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func optFromURL(u string) *redis.Options {
	opt, err := redis.ParseURL(u)
	die(err, "redis url")
	return opt
}
