package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/coder/websocket"
	"github.com/joho/godotenv"
)

var (
	clients = make(map[*websocket.Conn]bool)
	mu      sync.Mutex
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}

	mu.Lock()
	if len(clients) >= 2 {
		mu.Unlock()
		conn.Close(websocket.StatusPolicyViolation, "only 2 clients allowed")
		return
	}
	clients[conn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			return
		}

		mu.Lock()
		for c := range clients {
			if c != conn {
				c.Write(context.Background(), websocket.MessageText, msg)
			}
		}
		mu.Unlock()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ws", wsHandler)

	log.Printf("2-client test server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
