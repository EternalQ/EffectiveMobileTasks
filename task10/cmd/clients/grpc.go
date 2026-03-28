package clients

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/EternalQ/eff-mobile-tasks/task10/proto"
)

// Generated
type GrpcClient struct {
	client pb.UserServiceClient
	ctx    context.Context
}

func NewGrpcClient(addr string) (*GrpcClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcClient{
		client: pb.NewUserServiceClient(conn),
		ctx:    context.Background(),
	}, nil
}

func (c *GrpcClient) Run() {
	for {
		fmt.Println("\n1. Создать")
		fmt.Println("2. Обновить")
		fmt.Println("3. Удалить")
		fmt.Println("4. Список")
		fmt.Println("5. Вход")
		fmt.Println("0. Выход")

		choice := ReadInput("\nВыбор: ")

		switch choice {
		case "1":
			c.createUser()
		case "2":
			c.updateUser()
		case "3":
			c.deleteUser()
		case "4":
			c.listUsers()
		case "5":
			c.login()
		case "0":
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

func (c *GrpcClient) createUser() {
	name := ReadInput("Имя: ")
	if name == "" {
		fmt.Println("Имя не может быть пустым")
		return
	}

	email := ReadInput("Email: ")
	if email == "" {
		fmt.Println("Email не может быть пустым")
		return
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	user, err := c.client.CreateUser(ctx, &pb.CreateUserRequest{Name: name, Email: email})
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("✅ Создан: ID=%d, Name=%s, Email=%s\n", user.Id, user.Name, user.Email)
}

func (c *GrpcClient) updateUser() {
	id := ReadInt("ID: ")
	if id == 0 {
		fmt.Println("ID не может быть пустым")
		return
	}

	name := ReadInput("Новое имя (Enter - пропустить): ")
	email := ReadInput("Новый email (Enter - пропустить): ")

	if name == "" && email == "" {
		fmt.Println("Укажите хотя бы одно поле")
		return
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	user, err := c.client.UpdateUser(ctx, &pb.UpdateUserRequest{Id: id, Name: name, Email: email})
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("✅ Обновлен: ID=%d, Name=%s, Email=%s\n", user.Id, user.Name, user.Email)
}

func (c *GrpcClient) deleteUser() {
	id := ReadInt("ID для удаления: ")
	if id == 0 {
		fmt.Println("ID не может быть пустым")
		return
	}

	confirm := ReadInput(fmt.Sprintf("Удалить ID=%d? (y/N): ", id))
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Отменено")
		return
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("✅ Удален пользователь ID=%d\n", id)
}

func (c *GrpcClient) listUsers() {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	list, err := c.client.GetUsers(ctx, &emptypb.Empty{})
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if len(list.Users) == 0 {
		fmt.Println("Нет пользователей")
		return
	}

	fmt.Printf("\nВсего: %d\n", len(list.Users))
	fmt.Println(strings.Repeat("-", 40))
	for _, u := range list.Users {
		fmt.Printf("ID: %d | Имя: %s | Email: %s\n", u.Id, u.Name, u.Email)
	}
	fmt.Println(strings.Repeat("-", 40))
}

func (c *GrpcClient) login() {
	name := ReadInput("Имя: ")
	if name == "" {
		fmt.Println("Укажите имя")
		return
	}

	res, err := c.client.Login(c.ctx, &pb.LoginRequest{Name: name})
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	c.ctx = metadata.AppendToOutgoingContext(c.ctx, "authorization", "Bearer "+res.Token)
	fmt.Println("Залогинен")
}
