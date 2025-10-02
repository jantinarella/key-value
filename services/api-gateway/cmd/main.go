package main

import (
	"context"
	"key-value/client"
	"key-value/services/api-gateway/internal/config"
	"key-value/services/api-gateway/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	// Load configuration
	config := config.Load()

	// Create a new Echo instance
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	// Create a new KVStoreClient
	kvstoreClient, err := client.NewKVStoreClient(config.KVServiceAddr)
	if err != nil {
		e.Logger.Fatal("Failed to create KVStoreClient: %v", err)

	}
	defer kvstoreClient.Close()

	// Use middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Setup routes
	err = router.SetupRoutes(e, config, kvstoreClient)
	if err != nil {
		e.Logger.Fatal("Failed to setup routes: ", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server with graceful shutdown
	e.Logger.Info("🚀 Starting server on port " + config.Port)
	go func(port string) {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("❌ Server failed to start:", err)
		}
	}(config.Port)
	e.Logger.Info("✅ Server started successfully. Press Ctrl+C to shutdown gracefully.")

	// Wait for interrupt or kill signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	e.Logger.Info("\n🛑 Received shutdown signal, starting graceful shutdown...")

	// Create shutdown context with timeout and show countdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Show countdown in a goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for i := 10; i > 0; i-- {
			select {
			case <-shutdownCtx.Done():
				return
			case <-ticker.C:
				e.Logger.Info("⏰ Shutdown in", i, "seconds...")
			}
		}
	}()

	e.Logger.Info("⏳ Graceful shutdown initiated (10 second timeout)...")
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Error("❌ Server forced to shutdown:", err)
	} else {
		e.Logger.Info("✅ Server exited gracefully")
	}
}
