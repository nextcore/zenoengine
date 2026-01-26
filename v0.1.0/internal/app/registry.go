package app

import (
	"zeno/internal/slots"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/worker"

	"github.com/go-chi/chi/v5"
)

// RegisterAllSlots membungkus pendaftaran seluruh slot yang tersedia di ZenoEngine
func RegisterAllSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager, queue worker.JobQueue, setConfig func([]string)) {
	slots.RegisterRouterSlots(eng, r)
	slots.RegisterDBSlots(eng, dbMgr)
	slots.RegisterRawDBSlots(eng, dbMgr)
	slots.RegisterUtilSlots(eng)
	slots.RegisterBladeSlots(eng)   // Enable Blade Support
	slots.RegisterInertiaSlots(eng) // Enable Inertia.js Support
	slots.RegisterFileSystemSlots(eng)
	slots.RegisterValidatorSlots(eng)
	slots.RegisterAuthSlots(eng, dbMgr)
	slots.RegisterImageSlots(eng)
	slots.RegisterLogicSlots(eng)
	slots.RegisterSecuritySlots(eng)
	slots.RegisterNetworkSlots(eng)
	slots.RegisterHTTPServerSlots(eng)
	slots.RegisterMailSlots(eng)
	slots.RegisterCacheSlots(eng, nil)
	slots.RegisterJSONSlots(eng)
	slots.RegisterMathSlots(eng)
	slots.RegisterJobSlots(eng, queue, setConfig)
	slots.RegisterUploadSlots(eng)
	slots.RegisterCollectionSlots(eng)
	slots.RegisterTimeSlots(eng)
	slots.RegisterExcelSlots(eng)
	slots.RegisterHTTPClientSlots(eng)
	slots.RegisterFunctionSlots(eng)
}
