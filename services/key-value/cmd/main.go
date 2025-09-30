package main

import (
	"key-value/proto/keyvalue"
	"key-value/services/key-value/internal/config"
	"key-value/services/key-value/internal/kvstore"
	"key-value/services/key-value/internal/server"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	config := config.Load()

	// Create the key-value store
	store := kvstore.NewInMemoryStore()

	// Create the gRPC server
	grpcServer := grpc.NewServer()

	reflection.Register(grpcServer) // Allows for gRPC endpoit discovery (helpful for postman testing)

	// Create and register our service
	kvServer := server.NewKeyValueServer(store)
	keyvalue.RegisterKeyValueServiceServer(grpcServer, kvServer)

	lis, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", config.Port, err)
	}

	log.Printf("🚀 gRPC Key-Value server starting on port %s", config.Port)

	// Start server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	log.Println("✅ gRPC server started successfully. Press Ctrl+C to shutdown gracefully.")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Received shutdown signal, starting graceful shutdown...")

	// Graceful shutdown
	grpcServer.GracefulStop()
	log.Println("✅ gRPC server exited gracefully")
}
