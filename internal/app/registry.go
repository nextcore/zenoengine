package app

import (
	"zeno/internal/slots"
	pkgslots "zeno/pkg/slots"
	"zeno/pkg/blade"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/worker"

	"github.com/go-chi/chi/v5"
)

// RegisterAllSlots membungkus pendaftaran seluruh slot yang tersedia di ZenoEngine
func RegisterAllSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager, queue worker.JobQueue, setConfig func([]string)) {
	RegisterCoreSlots(eng)
	RegisterWebSlots(eng, r)
	RegisterDataSlots(eng, dbMgr)
	RegisterExtraSlots(eng, r, dbMgr, queue, setConfig)
}

// RegisterCoreSlots mendaftarkan slot dasar (Logic, Math, Time, dll)
func RegisterCoreSlots(eng *engine.Engine) {
	slots.RegisterUtilSlots(eng)
	slots.RegisterSecuritySlots(eng)
	pkgslots.RegisterLogicSlots(eng)
	slots.RegisterMathSlots(eng)
	slots.RegisterTimeSlots(eng)
	slots.RegisterFunctionSlots(eng)
	slots.RegisterMetaSlots(eng)
	slots.RegisterFileSystemSlots(eng)
	slots.RegisterStorageSlots(eng)
	slots.RegisterCollectionSlots(eng)
}

// RegisterWebSlots mendaftarkan slot untuk kebutuhan web (Routing, View, Session, dll)
func RegisterWebSlots(eng *engine.Engine, r *chi.Mux) {
	slots.RegisterRouterSlots(eng, r)
	blade.RegisterBladeSlots(eng)
	slots.RegisterInertiaSlots(eng)
	slots.RegisterHTTPServerSlots(eng)
	slots.RegisterSessionSlots(eng)
	slots.RegisterCaptchaSlots(eng, r)
	slots.RegisterUploadSlots(eng)
	slots.RegisterHTTPClientSlots(eng)
}

// RegisterDataSlots mendaftarkan slot untuk manipulasi data (DB, JSON, dll)
func RegisterDataSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {
	slots.RegisterDBSlots(eng, dbMgr)
	slots.RegisterRawDBSlots(eng, dbMgr)
	slots.RegisterSchemaSlots(eng, dbMgr)
	slots.RegisterORMSlots(eng, dbMgr)
	slots.RegisterValidatorSlots(eng, dbMgr)
	slots.RegisterAuthSlots(eng, dbMgr)
	slots.RegisterAspNetSlots(eng, dbMgr)
	slots.RegisterJSONSlots(eng)
	slots.RegisterDBHookSlots(eng)
}

// RegisterExtraSlots mendaftarkan slot tambahan yang opsional atau berat
func RegisterExtraSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager, queue worker.JobQueue, setConfig func([]string)) {
	slots.RegisterMailSlots(eng)
	slots.RegisterCacheSlots(eng, nil)
	slots.RegisterJobSlots(eng, queue, setConfig)
	if r != nil {
		slots.RegisterContainerBridgeSlots(eng, r)
	}
}
