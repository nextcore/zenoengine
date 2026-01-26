package livereload

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Broker manages SSE connections
type Broker struct {
	mu          sync.RWMutex
	clients     map[chan string]bool
	broadcaster chan string
}

// Global instance
var Instance *Broker

func init() {
	Instance = NewBroker()
	go Instance.Start()
}

func NewBroker() *Broker {
	return &Broker{
		clients:     make(map[chan string]bool),
		broadcaster: make(chan string),
	}
}

func (b *Broker) Start() {
	for {
		msg := <-b.broadcaster
		b.mu.RLock()
		for clientChan := range b.clients {
			select {
			case clientChan <- msg:
			default:
				// Skip if client is blocked
			}
		}
		b.mu.RUnlock()
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string)

	b.mu.Lock()
	b.clients[clientChan] = true
	b.mu.Unlock()

	// Cleanup on disconnect
	defer func() {
		b.mu.Lock()
		delete(b.clients, clientChan)
		close(clientChan)
		b.mu.Unlock()
	}()

	// Send initial connection message
	fmt.Fprintf(w, "data: connected\n\n")
	w.(http.Flusher).Flush()

	// Keep connection open
	notify := r.Context().Done()
	
	// Heartbeat ticker to keep connection alive
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			w.(http.Flusher).Flush()
		case <-notify:
			return
		}
	}
}

// Broadcast sends a reload signal to all connected clients
func Broadcast() {
	if Instance != nil {
		slog.Info("⚡ Sending Live Reload signal...")
		select {
		case Instance.broadcaster <- "reload":
		case <-time.After(1 * time.Second):
			slog.Warn("⚠️  Live Reload broadcast timed out")
		}
	}
}
