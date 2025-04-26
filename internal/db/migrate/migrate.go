package migrate

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_time TIMESTAMPTZ DEFAULT now(),
    url TEXT,
    referrer TEXT,
    user_agent TEXT,
    lang TEXT,
    screen_w INT,
    screen_h INT,
    session_id TEXT,
    hashed_ip TEXT,
    country TEXT,
    region TEXT,
    city TEXT
);
CREATE INDEX IF NOT EXISTS idx_event_api_key_time ON analytics_events(api_key, event_time DESC);
`)
	if err != nil {
		panic(err)
	}
}
