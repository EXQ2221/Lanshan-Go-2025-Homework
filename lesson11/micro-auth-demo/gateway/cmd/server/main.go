package main

import (
	"log"
	"os"

	"example.com/micro-auth-demo/gateway/internal/router"
	"example.com/micro-auth-demo/gateway/internal/rpc"
)

func main() {
	if err := rpc.InitAuthClient(getenv("AUTH_SERVICE_ADDR", "127.0.0.1:9002")); err != nil {
		log.Fatal(err)
	}
	if err := rpc.InitUserClient(getenv("USER_SERVICE_ADDR", "127.0.0.1:9001")); err != nil {
		log.Fatal(err)
	}

	addr := ":" + getenv("PORT", "8080")
	log.Printf("gateway listening on %s", addr)

	if err := router.New().Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
