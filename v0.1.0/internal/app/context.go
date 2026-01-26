package app

import (
	"net/http"
	"sync"
	"zeno/pkg/dbmanager"
	"zeno/pkg/worker"

	"github.com/go-chi/chi/v5"
)

// HotRouter adalah wrapper agar kita bisa mengganti router saat runtime
type HotRouter struct {
	mu     sync.RWMutex
	router *chi.Mux
}

func (h *HotRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.router == nil {
		http.Error(w, "ZenoEngine: Router not ready", http.StatusServiceUnavailable)
		return
	}
	h.router.ServeHTTP(w, r)
}

func (h *HotRouter) Swap(newRouter *chi.Mux) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.router = newRouter
}

// AppContext menyimpan dependensi global
type AppContext struct {
	DBMgr        *dbmanager.DBManager
	Queue        worker.JobQueue
	Env          string
	Hot          *HotRouter
	WorkerQueues []string
}
