package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"example.com/micro-auth-demo/user-service/internal/biz"
	"example.com/micro-auth-demo/user-service/internal/dal/mysql"
	"example.com/micro-auth-demo/user-service/internal/handler"
	"example.com/micro-auth-demo/user-service/internal/repository"
	"example.com/micro-auth-demo/user-service/kitex_gen/user/userservice"
	"github.com/cloudwego/kitex/server"
)

func main() {
	ctx := context.Background()
	rpcAddr := ":" + getenv("PORT", "9001")
	healthAddr := ":" + getenv("HEALTH_PORT", "19001")
	mysqlDSN := getenv("MYSQL_DSN", "demo:demo@tcp(127.0.0.1:3306)/micro_auth_demo?charset=utf8mb4&parseTime=True&loc=Local")

	db, err := mysql.Init(mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewUserRepository(db)
	service := biz.NewUserService(repo)
	if err := service.SeedDemoUser(ctx); err != nil {
		log.Fatal(err)
	}

	go serveHealth(healthAddr)

	addr, err := net.ResolveTCPAddr("tcp", rpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	svr := userservice.NewServer(
		handler.NewUserServiceImpl(service),
		server.WithServiceAddr(addr),
	)

	log.Printf("user-service kitex listening on %s", rpcAddr)
	if err := svr.Run(); err != nil {
		log.Fatal(err)
	}
}

func serveHealth(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("user-service health listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
