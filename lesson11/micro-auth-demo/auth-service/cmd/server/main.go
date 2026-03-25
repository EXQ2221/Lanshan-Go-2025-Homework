package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"example.com/micro-auth-demo/auth-service/internal/biz"
	mysqlstore "example.com/micro-auth-demo/auth-service/internal/dal/mysql"
	redisstore "example.com/micro-auth-demo/auth-service/internal/dal/redis"
	"example.com/micro-auth-demo/auth-service/internal/handler"
	"example.com/micro-auth-demo/auth-service/internal/repository"
	"example.com/micro-auth-demo/auth-service/internal/rpc"
	"example.com/micro-auth-demo/auth-service/kitex_gen/auth/authservice"
	"github.com/cloudwego/kitex/server"
)

func main() {
	rpcAddr := ":" + getenv("PORT", "9002")
	healthAddr := ":" + getenv("HEALTH_PORT", "19002")
	mysqlDSN := getenv("MYSQL_DSN", "demo:demo@tcp(127.0.0.1:3306)/micro_auth_demo?charset=utf8mb4&parseTime=True&loc=Local")
	redisAddr := getenv("REDIS_ADDR", "127.0.0.1:6379")
	userServiceAddr := getenv("USER_SERVICE_ADDR", "127.0.0.1:9001")
	jwtSecret := getenv("JWT_SECRET", "demo-secret")

	db, err := mysqlstore.Init(mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := redisstore.Init(redisAddr)
	if err != nil {
		log.Fatal(err)
	}

	userClient, err := rpc.NewUserClient(userServiceAddr)
	if err != nil {
		log.Fatal(err)
	}

	service := biz.NewAuthService(
		repository.NewSessionRepository(db),
		repository.NewRefreshTokenRepository(db),
		repository.NewSecurityEventRepository(db),
		repository.NewAuthCache(redisClient),
		userClient,
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
	)

	go serveHealth(healthAddr)

	addr, err := net.ResolveTCPAddr("tcp", rpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	svr := authservice.NewServer(
		handler.NewAuthServiceImpl(service),
		server.WithServiceAddr(addr),
	)

	log.Printf("auth-service kitex listening on %s", rpcAddr)
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

	log.Printf("auth-service health listening on %s", addr)
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
