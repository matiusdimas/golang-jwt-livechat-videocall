package handlers

import (
	
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// menyimpan koneksi dan username
var (
	clients = make(map[*websocket.Conn]string)
	mutex   = sync.Mutex{}
)

// setup upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // boleh disesuaikan untuk keamanan
	},
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	username := "anonymous"

	// ambil username dari context jika tersedia (hasil verifikasi JWT)
	if u := r.Context().Value("username"); u != nil {
		username = u.(string)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade WebSocket Failed:", err)
		return
	}
	defer conn.Close()

	// simpan koneksi ke dalam map
	mutex.Lock()
	clients[conn] = username
	mutex.Unlock()

	if username != "anonymous" {
		log.Printf("User %s join\n", username)
	}

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("User %s logout: %v\n", username, err)
			break
		}

		// abaikan pesan dari anonymous
		if username == "anonymous" {
			continue
		}

		msg := Message{
			Sender:  username,
			Message: string(msgBytes),
		}

		broadcast(msg)
	}

	// hapus koneksi saat disconnect
	mutex.Lock()
	delete(clients, conn)
	mutex.Unlock()
}

// kirim pesan ke semua client yang masih aktif
func broadcast(msg Message) {
	data, _ := json.Marshal(msg)

	mutex.Lock()
	defer mutex.Unlock()

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("Failed Sent to Client:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
