package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"
	"zeno/pkg/engine"
)

// Struktur Data Tugas
type JobPayload struct {
	ScriptPath string                 `json:"script_path"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Fungsi Utama Worker (Berjalan di Background)
func Start(ctx context.Context, eng *engine.Engine, queue JobQueue, queues []string) {
	// 1. CEK: Jika Queue Nil, matikan worker
	if queue == nil {
		slog.Info("🚫 Worker disabled: Queue not available")
		return
	}

	slog.Info("👷 Background Worker Started", "queues", queues)

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			slog.Info("👷 Worker Stopping... Waiting for active jobs to finish")
			wg.Wait()
			slog.Info("👷 Worker Fully Stopped")
			return
		default:
			// 1. Ambil Tugas dari Queue (Blocking/Polling)
			queueName, payloadBytes, err := queue.Pop(ctx, queues)

			if err != nil {
				// Cek jika error bukan context cancelled
				if ctx.Err() != nil {
					slog.Info("👷 Worker Stopping... Waiting for active jobs to finish")
					wg.Wait()
					slog.Info("👷 Worker Fully Stopped")
					return
				}
				// If error, log and retry
				slog.Error("❌ Worker Queue Error", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			payloadStr := string(payloadBytes)

			slog.Info("⚡ New Job Received", "queue", queueName)

			// 2. Parse Payload
			var job JobPayload
			if err := json.Unmarshal([]byte(payloadStr), &job); err != nil {
				slog.Error("❌ Invalid Job Payload", "error", err)
				continue
			}

			// 3. Eksekusi Script Zenolang
			wg.Add(1)
			go func() {
				defer wg.Done()
				executeJob(eng, job)
			}()
		}
	}
}

func executeJob(eng *engine.Engine, job JobPayload) {
	start := time.Now()

	// Load Script
	root, err := engine.LoadScript(job.ScriptPath)
	if err != nil {
		slog.Error("❌ Job Script Not Found", "path", job.ScriptPath, "error", err)
		return
	}

	// Siapkan Scope (Inject Data)
	scope := engine.NewScope(nil)
	for k, v := range job.Data {
		scope.Set(k, v)
	}

	// Set variabel standar job
	scope.Set("job_created_at", job.CreatedAt)

	// Execute
	// Gunakan Background Context (agar tidak putus jika request http putus)
	ctx := context.Background()
	if err := eng.Execute(ctx, root, scope); err != nil {
		slog.Error("❌ Job Failed", "path", job.ScriptPath, "error", err)
	} else {
		slog.Info("✅ Job Completed", "path", job.ScriptPath, "duration", time.Since(start))
	}
}
