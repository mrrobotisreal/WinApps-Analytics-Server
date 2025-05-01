package db

import (
	"WinApps-Analytics-Server/internal/config"
	"WinApps-Analytics-Server/internal/db/migrate"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log"
)

func InitPostgres(ctx context.Context, cfg config.Config) *pgxpool.Pool {
	pg, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("pg connection error: %v", err)
	}
	migrate.Run(ctx, pg)
	return pg
}

func InitRedis(ctx context.Context, cfg config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL[8:]})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	return rdb
}

//func Init(ctx context.Context, cfg config.Config) (*pgxpool.Pool, *redis.Client) {
//	pg, err := pgxpool.New(ctx, cfg.PostgresURL)
//	if err != nil {
//		log.Fatalf("pg connection error: %v", err)
//	}
//
//	migrate.Run(ctx, pg)
//
//	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL[8:]})
//	if err := rdb.Ping(ctx).Err(); err != nil {
//		log.Fatalf("redis connection error: %v", err)
//	}
//
//	return pg, rdb
//}
