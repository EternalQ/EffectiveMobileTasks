package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/EternalQ/eff-mobile-tasks/task10/internal/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *RestService) handleWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	defer ws.Close()

	if err := ws.WriteMessage(websocket.TextMessage, []byte("usage: /add <name>")); err != nil {
		log.Println("Ошибка первой записи:", err)
		return
	}
	for {
		t, p, err := ws.ReadMessage()
		if err != nil {
			log.Println("Ошибка чтения:", err)
			break
		}
		log.Printf("Получено: %s", p)

		msg := string(p)
		if strings.HasPrefix(msg, "/add ") {
			name := strings.TrimPrefix(msg, "/add ")
			if name != "" {
				s.register(&models.User{Name: name})
				continue
			}
		}

		p = []byte(fmt.Sprintf("echo: %s", p))
		if err := ws.WriteMessage(t, p); err != nil {
			log.Println("Ошибка записи:", err)
			break
		}
	}
}
