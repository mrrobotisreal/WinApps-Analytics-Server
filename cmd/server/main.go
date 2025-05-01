package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"WinApps-Analytics-Server/internal/config"
	"WinApps-Analytics-Server/internal/db"
	"WinApps-Analytics-Server/internal/handlers"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	//pg, redis := db.Init(context.Background(), cfg)
	//
	//go worker.StartFlusher(context.Background(), pg, redis, cfg.FlushInterval)

	pg := db.InitPostgres(context.Background(), cfg)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	{
		//api.POST("/event", handlers.NewEventHandler(cfg, redis))
		api.GET("/events", handlers.DashboardHandler(pg))
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	log.Printf("starting HTTPS server on port %s", cfg.Port)
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
