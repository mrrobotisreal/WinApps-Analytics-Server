package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func DashboardHandler(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 3*time.Second)
		defer cancel()
		rows, err := db.Query(ctx, `SELECT event_type, count(*) FROM analytics_events WHERE event_time > now() - interval '24 hours' GROUP BY event_type`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		result := make(map[string]int64)
		for rows.Next() {
			var et string
			var cnt int64
			_ = rows.Scan(&et, &cnt)
			result[et] = cnt
		}
		c.JSON(http.StatusOK, result)
	}
}
