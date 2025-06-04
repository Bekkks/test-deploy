package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	// Добавляем клиента
	clients[ws] = true

	// Просим имя
	err = ws.WriteMessage(websocket.TextMessage, []byte("Привет! Введите ваше имя:"))
	if err != nil {
		log.Println("Write error:", err)
		delete(clients, ws)
		return
	}

	// Читаем имя
	_, nameMsg, err := ws.ReadMessage()
	if err != nil {
		log.Println("Read name error:", err)
		ws.WriteMessage(websocket.TextMessage, []byte("Ошибка при получении имени"))
		delete(clients, ws)
		return
	}

	name := strings.TrimSpace(strings.ToLower(string(nameMsg)))
	if name != "beka" {
		ws.WriteMessage(websocket.TextMessage, []byte("Неправильное имя пользователя"))
		delete(clients, ws)
		return
	}

	// Приветствие
	ws.WriteMessage(websocket.TextMessage, []byte("Добро пожаловать в чат, Beka!"))

	// Обработка сообщений
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Read message error:", err)
			delete(clients, ws)
			break
		}
		broadcast <- "Beka: " + string(msg)
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}
