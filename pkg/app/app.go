package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/nextcore/zeno-go/pkg/engine"
	internalapp "github.com/nextcore/zenoengine/internal/app"
	"github.com/nextcore/zenoengine/pkg/dbmanager"
	"github.com/nextcore/zenoengine/pkg/worker"
)

// RegisterAllSlots registers all available default slots in ZenoEngine for external modules.
func RegisterAllSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager, queue worker.JobQueue, setConfig func([]string)) {
	internalapp.RegisterAllSlots(eng, r, dbMgr, queue, setConfig)
}
