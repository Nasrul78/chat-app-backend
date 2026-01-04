package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
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
	http.HandleFunc("/ws", wsHandler)

	log.Println("2-client test server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
