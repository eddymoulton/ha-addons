package main

import (
	"embed"
	"ha-config-history/internal/api"
	"ha-config-history/internal/core"
	"ha-config-history/internal/types"

	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	version   = "dev"
	gitCommit = "none"
	startTime time.Time
)

//go:embed dist
var distFS embed.FS

//go:embed dist/_app
var appFS embed.FS

func main() {
	startTime = time.Now()

	// Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	config := types.LoadConfig("config.json")

	server := core.NewServer(config)
	server.Start()

	r := gin.New()

	// Request ID middleware
	r.Use(func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	r.Use(gin.Recovery())

	fs, err := static.EmbedFolder(distFS, "dist")
	if err != nil {
		slog.Error("Failed to embed static files", "error", err)
		os.Exit(1)
	}
	r.Use(static.Serve("/", fs))

	appFs, err := static.EmbedFolder(appFS, "dist/_app")
	if err != nil {
		slog.Error("Failed to embed app static files", "error", err)
		os.Exit(1)
	}
	r.Use(static.Serve("/_app", appFs))

	r.GET("/configs", api.GetConfigsHandler(server))
	r.GET("/configs/:group/:id/backups", api.ListConfigBackupsHandler(server))
	r.GET("/configs/:group/:id/backups/:filename", api.GetConfigBackupHandler(server))
	r.GET("/configs/:group/:id/compare/:left/diff/:right", api.GetBackupDiffHandler(server))
	r.POST("/configs/:group/:id/backups/:filename/restore", api.RestoreBackupHandler(server))
	r.DELETE("/configs/:group/:id/backups/:filename", api.DeleteConfigBackupHandler(server))
	r.DELETE("/configs/:group/:id", api.DeleteAllConfigBackupsHandler(server))
	r.POST("/backup", api.ProcessConfigsHandler(server))
	r.GET("/settings", api.GetSettingsHandler(server))
	r.PUT("/settings", api.UpdateSettingsHandler(server))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"uptime": time.Since(startTime).String(),
		})
	})

	// Version endpoint
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": version,
			"commit":  gitCommit,
		})
	})

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		server.Shutdown()
		os.Exit(0)
	}()

	slog.Info("Starting API server", "port", config.Port)
	if err := r.Run(config.Port); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
