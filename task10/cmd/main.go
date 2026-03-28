package main

import (
	"fmt"
	"log"
	"os"
	"time"

	clients "github.com/EternalQ/eff-mobile-tasks/task10/cmd/clients"
	"github.com/EternalQ/eff-mobile-tasks/task10/pkg/rabbit"
	"github.com/joho/godotenv"
)

var (
	grpcPort  string
	restPort  string
	rmqHost   string
	notifExch string
)

func load() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	grpcPort = os.Getenv("GRPC_PORT")
	restPort = os.Getenv("REST_PORT")
	rmqHost = os.Getenv("RMQ_HOST")
	notifExch = os.Getenv("EXCHANGE")
}

func main() {
	load()

	fmt.Println("\n1. GRPC")
	fmt.Println("2. WS")
	fmt.Println("3. Rabbit sub")
	fmt.Println("4. REST")
	fmt.Println("0. Выход")

	choice := clients.ReadInput("\nВыбор: ")

	switch choice {
	case "1":
		c, err := clients.NewGrpcClient("localhost:" + grpcPort)
		if err != nil {
			log.Fatalf("Ошибка подключения: %v", err)
		}
		c.Run()
	case "2":
		clients.NewWsClient("localhost:" + restPort).Run()
	case "3":
		c := rabbit.NewClient(rmqHost, notifExch)
		msgs, err := c.Sub()
		if err != nil {
			log.Fatalf("Ошибка подключения: %v", err)
		}
		for {
			d := <-msgs
			fmt.Printf("[%v] Notif: %s\n", d.Timestamp, d.Body)
			time.Sleep(100 * time.Millisecond)
		}
	case "4":
		c := clients.NewRestClient("localhost:" + restPort)
		c.Run()
	case "0":
		fmt.Println("До свидания!")
		return
	default:
		fmt.Println("Неверный выбор")
	}
}
