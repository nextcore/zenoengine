package app

import (
	"github.com/nextcore/zenoengine/internal/slots"
	pkgslots "github.com/nextcore/zeno-go/pkg/slots"
	"github.com/nextcore/zeno-go/pkg/blade"
	"github.com/nextcore/zenoengine/pkg/dbmanager"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zenoengine/pkg/worker"

	"github.com/go-chi/chi/v5"
)

// registerConfig menyimpan konfigurasi modular untuk pendaftaran slot
type registerConfig struct {
	// Core Slots
	util       bool
	security   bool
	logic      bool
	math       bool
	time       bool
	function   bool
	meta       bool
	fs         bool
	storage    bool
	collection bool

	// Web Slots
	router             bool
	routerMux          *chi.Mux
	blade              bool
	inertia            bool
	httpServer         bool
	session            bool
	captcha            bool
	upload             bool
	httpClient         bool

	// Data Slots
	dbMgr     *dbmanager.DBManager
	db        bool
	rawDb     bool
	schema    bool
	orm       bool
	validator bool
	auth      bool
	aspnet    bool
	json      bool
	dbHook    bool

	// Extra Slots
	mail            bool
	cache           bool
	job             bool
	containerBridge bool
	queue           worker.JobQueue
	setConfig       func([]string)
}

// RegisterOption mendefinisikan tanda tangan fungsi opsi konfigurasi
type RegisterOption func(*registerConfig)

// WithCore mendaftarkan seluruh slot dasar (Core Slots) secara kolektif
func WithCore() RegisterOption {
	return func(c *registerConfig) {
		c.util = true
		c.security = true
		c.logic = true
		c.math = true
		c.time = true
		c.function = true
		c.meta = true
		c.fs = true
		c.storage = true
		c.collection = true
	}
}

// WithWeb mendaftarkan seluruh slot web secara kolektif dengan router yang ditentukan
func WithWeb(r *chi.Mux) RegisterOption {
	return func(c *registerConfig) {
		c.router = true
		c.routerMux = r
		c.blade = true
		c.inertia = true
		c.httpServer = true
		c.session = true
		c.captcha = true
		c.upload = true
		c.httpClient = true
		c.containerBridge = true
	}
}

// WithData mendaftarkan seluruh slot pengolahan data secara kolektif dengan DBManager yang ditentukan
func WithData(dbMgr *dbmanager.DBManager) RegisterOption {
	return func(c *registerConfig) {
		c.dbMgr = dbMgr
		c.db = true
		c.rawDb = true
		c.schema = true
		c.orm = true
		c.validator = true
		c.auth = true
		c.aspnet = true
		c.json = true
		c.dbHook = true
	}
}

// WithExtra mendaftarkan slot tambahan secara kolektif (Mail, Cache, Jobs)
func WithExtra(queue worker.JobQueue, setConfig func([]string)) RegisterOption {
	return func(c *registerConfig) {
		c.mail = true
		c.cache = true
		c.job = true
		c.queue = queue
		c.setConfig = setConfig
	}
}

// OPSI GRANULAR (Untuk memilih fitur secara mendalam)

// WithBlade mengaktifkan pendaftaran slot Blade template
func WithBlade() RegisterOption {
	return func(c *registerConfig) {
		c.blade = true
	}
}

// WithRouter mengaktifkan pendaftaran slot web router
func WithRouter(r *chi.Mux) RegisterOption {
	return func(c *registerConfig) {
		c.router = true
		c.routerMux = r
	}
}

// WithDB mengaktifkan pendaftaran slot database (DB, ORM, Schema, DB Hook, dll)
func WithDB(dbMgr *dbmanager.DBManager) RegisterOption {
	return func(c *registerConfig) {
		c.dbMgr = dbMgr
		c.db = true
		c.rawDb = true
		c.schema = true
		c.orm = true
		c.dbHook = true
	}
}

// WithAuth mengaktifkan pendaftaran slot Auth & ASP.NET membership
func WithAuth(dbMgr *dbmanager.DBManager) RegisterOption {
	return func(c *registerConfig) {
		c.dbMgr = dbMgr
		c.auth = true
		c.aspnet = true
	}
}

// WithValidator mengaktifkan pendaftaran slot Validator
func WithValidator(dbMgr *dbmanager.DBManager) RegisterOption {
	return func(c *registerConfig) {
		c.dbMgr = dbMgr
		c.validator = true
	}
}

// WithMail mengaktifkan pendaftaran slot pengiriman Email
func WithMail() RegisterOption {
	return func(c *registerConfig) {
		c.mail = true
	}
}

// WithCache mengaktifkan pendaftaran slot in-memory caching
func WithCache() RegisterOption {
	return func(c *registerConfig) {
		c.cache = true
	}
}

// WithJob mengaktifkan pendaftaran slot background jobs/worker queue
func WithJob(queue worker.JobQueue, setConfig func([]string)) RegisterOption {
	return func(c *registerConfig) {
		c.job = true
		c.queue = queue
		c.setConfig = setConfig
	}
}

// RegisterSlots mendaftarkan slot ke Engine secara selektif berdasarkan opsi yang dipilih
func RegisterSlots(eng *engine.Engine, opts ...RegisterOption) {
	c := &registerConfig{}
	for _, opt := range opts {
		opt(c)
	}

	// 1. Core Slots
	if c.util {
		slots.RegisterUtilSlots(eng)
	}
	if c.security {
		slots.RegisterSecuritySlots(eng)
	}
	if c.logic {
		pkgslots.RegisterLogicSlots(eng)
	}
	if c.math {
		slots.RegisterMathSlots(eng)
	}
	if c.time {
		slots.RegisterTimeSlots(eng)
	}
	if c.function {
		slots.RegisterFunctionSlots(eng)
	}
	if c.meta {
		slots.RegisterMetaSlots(eng)
	}
	if c.fs {
		slots.RegisterFileSystemSlots(eng)
	}
	if c.storage {
		slots.RegisterStorageSlots(eng)
	}
	if c.collection {
		slots.RegisterCollectionSlots(eng)
	}

	// 2. Web Slots
	if c.router && c.routerMux != nil {
		slots.RegisterRouterSlots(eng, c.routerMux)
	}
	if c.blade {
		blade.RegisterBladeSlots(eng)
	}
	if c.inertia {
		slots.RegisterInertiaSlots(eng)
	}
	if c.httpServer {
		slots.RegisterHTTPServerSlots(eng)
	}
	if c.session {
		slots.RegisterSessionSlots(eng)
	}
	if c.captcha && c.routerMux != nil {
		slots.RegisterCaptchaSlots(eng, c.routerMux)
	}
	if c.upload {
		slots.RegisterUploadSlots(eng)
	}
	if c.httpClient {
		slots.RegisterHTTPClientSlots(eng)
	}

	// 3. Data Slots
	if c.db && c.dbMgr != nil {
		slots.RegisterDBSlots(eng, c.dbMgr)
	}
	if c.rawDb && c.dbMgr != nil {
		slots.RegisterRawDBSlots(eng, c.dbMgr)
	}
	if c.schema && c.dbMgr != nil {
		slots.RegisterSchemaSlots(eng, c.dbMgr)
	}
	if c.orm && c.dbMgr != nil {
		slots.RegisterORMSlots(eng, c.dbMgr)
	}
	if c.validator && c.dbMgr != nil {
		slots.RegisterValidatorSlots(eng, c.dbMgr)
	}
	if c.auth && c.dbMgr != nil {
		slots.RegisterAuthSlots(eng, c.dbMgr)
	}
	if c.aspnet && c.dbMgr != nil {
		slots.RegisterAspNetSlots(eng, c.dbMgr)
	}
	if c.json {
		slots.RegisterJSONSlots(eng)
	}
	if c.dbHook {
		slots.RegisterDBHookSlots(eng)
	}

	// 4. Extra Slots
	if c.mail {
		slots.RegisterMailSlots(eng)
	}
	if c.cache {
		slots.RegisterCacheSlots(eng, nil)
	}
	if c.job {
		slots.RegisterJobSlots(eng, c.queue, c.setConfig)
	}
	if c.containerBridge && c.routerMux != nil {
		slots.RegisterContainerBridgeSlots(eng, c.routerMux)
	}
}

// RegisterAllSlots membungkus pendaftaran seluruh slot yang tersedia di ZenoEngine (Backward Compatibility)
func RegisterAllSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager, queue worker.JobQueue, setConfig func([]string)) {
	RegisterSlots(eng,
		WithCore(),
		WithWeb(r),
		WithData(dbMgr),
		WithExtra(queue, setConfig),
	)
}
