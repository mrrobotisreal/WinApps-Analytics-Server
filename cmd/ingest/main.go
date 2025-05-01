package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"os"
	"time"
)

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
	brokers := getenv("KAFKA_BROKERS", "kafka:9092")
	topic := getenv("KAFKA_TOPIC", "analytics_events")
	addr := getenv("HTTP_ADDR", ":8080")
	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/api/event", func(c *gin.Context) {
		var ev Event
		if err := c.ShouldBindJSON(&ev); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if ev.TS == 0 {
			ev.TS = time.Now().UnixMilli()
		}
		b, _ := json.Marshal(ev)
		if err := writer.WriteMessages(
			context.Background(),
			kafka.Message{Key: []byte(ev.APIKey), Value: b, Time: time.Now()},
		); err != nil {
			log.Println("kafka write:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "kafka error"})
			return
		}
		c.Status(http.StatusAccepted)
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("ingest HTTPS listening on %s → Kafka %s/%s", addr, brokers, topic)
	log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
