package clients

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	u url.URL
}

func NewWsClient(host string) *WsClient {
	return &WsClient{url.URL{Scheme: "ws", Host: host, Path: "/ws"}}
}

func (c *WsClient) Run() {
	conn, _, err := websocket.DefaultDialer.Dial(c.u.String(), nil)
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	fmt.Println("\nadd user - /add <name>")

	for {
		msg := ReadInput("\nMsg: ")
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Fatalf("Ошибка отправки: %v", err)
		}

		_, res, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("Ошибка чтения: %v", err)
		}
		fmt.Printf("Res: %s\n", res)
	}
}
