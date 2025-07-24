package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type SignalMessage struct {
	Type    string          `json:"type"`    // "offer", "answer", "candidate"
	From    string          `json:"from"`    // pengirim
	To      string          `json:"to"`      // penerima
	Payload json.RawMessage `json:"payload"` // isi pesan signaling
}

var (
	videoClients = make(map[string]*websocket.Conn)
	videoMutex   = sync.Mutex{}
)

var videoUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func VideoCallWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("WS /ws-video handler dipanggil")
	username := "anonymous"

	// Ambil username dari JWT middleware context (pastikan middleware kirim "username")
	if u := r.Context().Value("username"); u != nil {
		username = u.(string)
	}

	conn, err := videoUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[Video WS] Gagal upgrade:", err)
		return
	}
	defer conn.Close()

	// Simpan koneksi user
	videoMutex.Lock()
	videoClients[username] = conn
	videoMutex.Unlock()

	log.Printf("[Video WS] %s terhubung", username)

	// Baca pesan dari user
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[Video WS] %s disconnect: %v", username, err)
			break
		}

		var msg SignalMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Println("[Video WS] Error parsing:", err)
			continue
		}

		msg.From = username // overwrite "from" biar tidak bisa dipalsukan

		// Kirim ke penerima
		videoMutex.Lock()
		receiverConn, ok := videoClients[msg.To]
		videoMutex.Unlock()

		if !ok {
			log.Printf("[Video WS] %s is not online", msg.To)
			continue
		}

		if err := receiverConn.WriteJSON(msg); err != nil {
			log.Printf("[Video WS] Failed Sent to %s: %v", msg.To, err)
			receiverConn.Close()
			videoMutex.Lock()
			delete(videoClients, msg.To)
			videoMutex.Unlock()
		}
	}

	// Hapus koneksi saat keluar
	videoMutex.Lock()
	delete(videoClients, username)
	videoMutex.Unlock()
	log.Printf("[Video WS] %s logout", username)
}
