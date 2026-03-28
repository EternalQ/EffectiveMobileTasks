package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/EternalQ/eff-mobile-tasks/task10/internal/service"
	"github.com/EternalQ/eff-mobile-tasks/task10/pkg/rabbit"
	pb "github.com/EternalQ/eff-mobile-tasks/task10/proto"
	"github.com/joho/godotenv"
)

var (
	grpcPort  string
	restPort  string
	secretKey string
	rmqHost   string
	notifExch string
)

func load() {
	godotenv.Load()

	grpcPort = os.Getenv("GRPC_PORT")
	restPort = os.Getenv("REST_PORT")
	secretKey = os.Getenv("SECRET_KEY")
	rmqHost = os.Getenv("RMQ_HOST")
	notifExch = os.Getenv("EXCHANGE")
}

func main() {
	load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	jwt := service.NewJWTManager(secretKey, 1*time.Hour)
	s := grpc.NewServer(grpc.UnaryInterceptor(jwt.AuthInterceptor()))
	pb.RegisterUserServiceServer(s, service.NewGrpcService(jwt))

	gl, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := s.Serve(gl); err != nil {
			panic(err)
		}
	}()

	rmq := rabbit.NewClient(rmqHost, notifExch)

	rest := service.NewRestService(logger, rmq).SetupRouter()
	http.ListenAndServe(":"+restPort, rest)
}
