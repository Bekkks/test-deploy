package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// В продакшене замените на конкретный домен
		return os.Getenv("ENV") == "development" || true
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex // Мьютекс для потокобезопасности
	broadcast = make(chan string)
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback порт
	}

	// Статические файлы
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Printf("Сервер запущен на :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка подключения:", err)
		return
	}

	// Регистрация клиента
	registerClient(ws)
	defer unregisterClient(ws)

	// Приветствие
	if err := ws.WriteMessage(websocket.TextMessage, []byte("Добро пожаловать в чат!")); err != nil {
		log.Println("Ошибка приветствия:", err)
		return
	}

	// Чтение сообщений
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Ошибка чтения:", err)
			break
		}
		broadcast <- string(msg)
	}
}

func registerClient(ws *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[ws] = true
	log.Println("Новый клиент подключён")
}

func unregisterClient(ws *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	ws.Close()
	delete(clients, ws)
	log.Println("Клиент отключён")
}

func handleMessages() {
	for msg := range broadcast {
		clientsMu.Lock()
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				unregisterClient(client)
			}
		}
		clientsMu.Unlock()
	}
}