package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sso-server/internal/database"
	"sso-server/internal/server"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(fiberServer *server.FiberServer, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := fiberServer.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {

	privKey, pubKey, err := loadKeys()
	if err != nil {
		log.Fatal("Could not load RSA KEYS", err)
	}
	server := server.New(privKey, pubKey)
	errDB := database.New().SeedPermissionsAndRoles()
	if errDB != nil {
		log.Fatal(err)
	}
	server.RegisterFiberRoutes()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)
	go func() {
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		err := server.Listen(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}

func loadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKeyData, err := os.ReadFile(os.Getenv("RSA_PRIVATE_KEY_PATH"))
	if err != nil {
		return nil, nil, err
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyData)
	if err != nil {
		return nil, nil, err
	}
	pubKeyData, err := os.ReadFile(os.Getenv("RSA_PUBLIC_KEY_PATH"))
	if err != nil {
		return nil, nil, err
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("Loaded keys from %s and %s\n", os.Getenv("RSA_PRIVATE_KEY_PATH"), os.Getenv("RSA_PUBLIC_KEY_PATH"))
	return privKey, pubKey, nil
}
