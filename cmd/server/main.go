// github.com/Dmitrii-Khramtsov/orderservice/cmd/server/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap"
)

	func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	app, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	app.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.Shutdown(ctx)
}
