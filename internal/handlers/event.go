package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"WinApps-Analytics-Server/internal/config"
)

type incomingEvent struct {
	EventType string `json:"event_type" binding:"required"`
	URL       string `json:"url" binding:"required,url"`
	Referrer  string `json:"referrer"`
	Lang      string `json:"lang"`
	ScreenW   int    `json:"screen_w"`
	ScreenH   int    `json:"screen_h"`
	SessionID string `json:"session_id"`
}

func NewEventHandler(cfg config.Config, redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ev incomingEvent
		if err := c.ShouldBindJSON(&ev); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c, 2*time.Second)
		defer cancel()
		if err := redis.RPush(ctx, "event_queue", c.MustGet(gin.BindKey).([]byte)).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "queue error"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
	}
}
