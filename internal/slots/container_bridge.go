package slots

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/go-chi/chi/v5"
)

// Circuit Breaker State
type circuitState struct {
	failCount  int
	lastFail   time.Time
	openUntil  time.Time
}

var (
	cbMu    sync.RWMutex
	cbState = make(map[string]*circuitState)
)

const (
	maxFailures      = 5
	resetTimeout     = 30 * time.Second
	failureThreshold = 5 * time.Minute
)

// Service Discovery Structs
type serviceNode struct {
	addr      string
	isHealthy bool
	lastCheck time.Time
	weight    int       // [NEW] Weighted LB
	expiresAt time.Time // [NEW] Auto-Pruning for Dynamic Nodes
}

type servicePool struct {
	nodes   []*serviceNode
	index   uint32 // for round-robin
	checkId string // unique check endpoint
	mu      sync.RWMutex
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]*servicePool)
)

// RegisterContainerBridgeSlots mendaftarkan slot khusus untuk memanggil Container lain
func RegisterContainerBridgeSlots(eng *engine.Engine, r chi.Router) {
	// Start Health Checker
	go startServiceHealthChecker()

	// Register Dynamic Discovery API if router is provided
	if r != nil {
		registerDiscoveryAPI(r)
	}

	// ==========================================
	// SLOT: DOCKER.HEALTH
	// ==========================================
	eng.Register("docker.health", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		host := "localhost"
		port := "8000"
		targetAs := "health"

		if node.Value != nil {
			host = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "port" {
				port = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "host" {
				host = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				targetAs = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		url := fmt.Sprintf("http://%s:%s/up", host, port)
		
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		
		result := map[string]interface{}{
			"host": host,
			"port": port,
		}

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			result["status"] = "healthy"
			resp.Body.Close()
		} else {
			result["status"] = "unhealthy"
			if err != nil {
				result["error"] = err.Error()
			} else {
				result["error"] = fmt.Sprintf("HTTP %d", resp.StatusCode)
				resp.Body.Close()
			}
		}

		scope.Set(targetAs, result)
		return nil
	}, engine.SlotMeta{
		Description: "Check health of a docker sidecar HTTP service.",
		Example:     "docker.health: 'php_worker' {\n  port: 8000\n  as: $h\n}",
		Inputs: map[string]engine.InputMeta{
			"host": {Description: "Hostname of the container", Required: false},
			"port": {Description: "Port of the container", Required: false},
			"as":   {Description: "Variable to store result", Required: false},
		},
	})

	// ==========================================
	// SLOT: DOCKER.CALL
	// ==========================================
	eng.Register("docker.call", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		host := coerce.ToString(resolveValue(node.Value, scope))
		if host == "" {
			return fmt.Errorf("docker.call requires a valid host/container name")
		}

		port := "80"
		endpoint := "/"
		method := "POST"
		var payload interface{}
		targetAs := "response"
		timeoutMs := 15000 // default 15s
		retryCount := 0
		useCB := false

		for _, c := range node.Children {
			if c.Name == "port" {
				port = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "endpoint" {
				endpoint = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "method" {
				method = strings.ToUpper(coerce.ToString(parseNodeValue(c, scope)))
			}
			if c.Name == "payload" {
				payload = parseNodeValue(c, scope)
			}
			if c.Name == "timeout" {
				timeoutMs, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
			if c.Name == "retry" {
				retryCount, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
			if c.Name == "circuit_breaker" {
				useCB, _ = coerce.ToBool(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				targetAs = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// 1. Resolve Service Discovery (Load Balancing)
		registryMu.RLock()
		pool, isService := registry[host]
		registryMu.RUnlock()

		if isService {
			healthyAddr := pool.getNextHealthyNode()
			if healthyAddr == "" {
				scope.Set(targetAs, map[string]interface{}{
					"success": false,
					"error":   "Service Discovery: No healthy nodes available for " + host,
					"host":    host,
				})
				return nil
			}
			host = healthyAddr
		}
		
		slog.Debug("🔌 Bridge calling", "host", host, "endpoint", endpoint)

		// 2. Check Circuit Breaker
		if useCB {
			cbMu.RLock()
			state, exists := cbState[host]
			cbMu.RUnlock()

			if exists && time.Now().Before(state.openUntil) {
				scope.Set(targetAs, map[string]interface{}{
					"success":         false,
					"error":           "Circuit Breaker: Open",
					"host":            host,
					"circuit_blocked": true,
				})
				return nil
			}
		}

		// Normalize endpoint
		if !strings.HasPrefix(endpoint, "/") {
			endpoint = "/" + endpoint
		}

		var url string
		if strings.Contains(host, ":") {
			url = fmt.Sprintf("http://%s%s", host, endpoint)
		} else {
			url = fmt.Sprintf("http://%s:%s%s", host, port, endpoint)
		}

		var resBodyBytes []byte
		var resCode int
		var finalErr error
		var isOk bool

		// Retry Loop
		for attempt := 0; attempt <= retryCount; attempt++ {
			var reqBody io.Reader
			if payload != nil {
				jsonData, err := json.Marshal(payload)
				if err != nil {
					return fmt.Errorf("docker.call payload encode error: %v", err)
				}
				reqBody = bytes.NewBuffer(jsonData)
			}

			req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
			if err != nil {
				return fmt.Errorf("docker.call request error: %v", err)
			}

			if payload != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("User-Agent", "ZenoEngine-ContainerBridge/1.1")

			client := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
			resp, err := client.Do(req)

			if err != nil {
				finalErr = err
				if attempt < retryCount {
					time.Sleep(200 * time.Millisecond) // exponential backoff could be better, but fixed for now
					continue
				}
				break
			}

			// Read Success
			body, _ := io.ReadAll(resp.Body)
			resBodyBytes = body
			resCode = resp.StatusCode
			isOk = resCode >= 200 && resCode < 300
			resp.Body.Close()

			if isOk || attempt == retryCount {
				break
			}
			// If not OK and have retries, try again
			time.Sleep(100 * time.Millisecond)
		}

		// 2. Update Circuit Breaker State
		if useCB {
			cbMu.Lock()
			state, exists := cbState[host]
			if !exists {
				state = &circuitState{}
				cbState[host] = state
			}

			if !isOk || finalErr != nil {
				state.failCount++
				state.lastFail = time.Now()
				if state.failCount >= maxFailures {
					state.openUntil = time.Now().Add(resetTimeout)
				}
			} else {
				// Reset on success
				state.failCount = 0
				state.openUntil = time.Time{}
			}
			cbMu.Unlock()
		}

		// 3. Final Result
		if finalErr != nil && resBodyBytes == nil {
			scope.Set(targetAs, map[string]interface{}{
				"success": false,
				"error":   finalErr.Error(),
				"host":    host,
			})
			return nil
		}

		// Try parsing JSON response
		var jsonResponse interface{}
		errJson := json.Unmarshal(resBodyBytes, &jsonResponse)

		result := map[string]interface{}{
			"success": isOk,
			"code":    resCode,
			"raw":     string(resBodyBytes),
		}

		if errJson == nil {
			result["data"] = jsonResponse
		}

		scope.Set(targetAs, result)
		return nil
	}, engine.SlotMeta{
		Description: "Call an external docker container microservice with resilience (retry & circuit breaker).",
		Example:     "docker.call: 'php_worker' {\n  endpoint: '/calculate'\n  payload: { data: 1 }\n  retry: 3\n  circuit_breaker: true\n  as: $res\n}",
		Inputs: map[string]engine.InputMeta{
			"port":            {Description: "Port (default 80)", Required: false},
			"endpoint":        {Description: "HTTP Path (default /)", Required: false},
			"method":          {Description: "HTTP Method (default POST)", Required: false},
			"payload":         {Description: "JSON array/object to send", Required: false},
			"timeout":         {Description: "Timeout in ms (default 15000)", Required: false},
			"retry":           {Description: "Number of retries on failure (default 0)", Required: false},
			"circuit_breaker": {Description: "Enable circuit breaker protection (default false)", Required: false},
			"as":              {Description: "Variable to store result", Required: false},
		},
	})

	// ==========================================
	// SLOT: DOCKER.NODES
	// ==========================================
	eng.Register("docker.nodes", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		serviceName := coerce.ToString(resolveValue(node.Value, scope))
		if serviceName == "" {
			return fmt.Errorf("docker.nodes requires a service name")
		}

		var addrs []string
		checkPath := "/health"
		weight := 100
		ttl := 0

		for _, c := range node.Children {
			if c.Name == "nodes" || c.Name == "val" {
				val := parseNodeValue(c, scope)
				if list, err := coerce.ToSlice(val); err == nil {
					for _, item := range list {
						addrs = append(addrs, coerce.ToString(item))
					}
				} else {
					// Support comma-separated string
					s := coerce.ToString(val)
					if strings.Contains(s, ",") {
						parts := strings.Split(s, ",")
						for _, p := range parts {
							addrs = append(addrs, strings.TrimSpace(p))
						}
					} else {
						addrs = append(addrs, s)
					}
				}
			}
			if c.Name == "check" {
				checkPath = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "weight" {
				weight, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
			if c.Name == "ttl" {
				ttl, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
		}

		if len(addrs) == 0 {
			return fmt.Errorf("docker.nodes: at least one node address is required")
		}

		registryMu.Lock()
		pool, exists := registry[serviceName]
		if !exists {
			pool = &servicePool{
				checkId: checkPath,
			}
			registry[serviceName] = pool
		}
		registryMu.Unlock()

		pool.mu.Lock()
		defer pool.mu.Unlock()

		var expiresAt time.Time
		if ttl > 0 {
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		}

		for _, addr := range addrs {
			// Basic normalization
			addr = strings.TrimPrefix(addr, "http://")
			addr = strings.TrimPrefix(addr, "https://")
			
			// Check if node exists to update it, or append
			found := false
			for _, n := range pool.nodes {
				if n.addr == addr {
					n.weight = weight
					n.expiresAt = expiresAt
					n.isHealthy = true
					found = true
					break
				}
			}

			if !found {
				pool.nodes = append(pool.nodes, &serviceNode{
					addr:      addr,
					isHealthy: true,
					weight:    weight,
					expiresAt: expiresAt,
				})
			}
		}

		slog.Info("📡 Service Registry updated", "service", serviceName, "nodes", len(addrs))
		return nil
	}, engine.SlotMeta{
		Description: "Register a pool of nodes for a service name (Load Balancing).",
		Example:     "docker.nodes: 'payment_service' {\n  nodes: ['10.0.0.1', '10.0.0.2']\n  check: '/status'\n}",
	})
}

// ==========================================
// HELPERS & BACKGROUND WORKERS
// ==========================================

func (p *servicePool) getNextHealthyNode() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.nodes) == 0 {
		return ""
	}

	// 1. Filter healthy and non-expired nodes
	var eligible []*serviceNode
	now := time.Now()
	for _, n := range p.nodes {
		if n.isHealthy && (n.expiresAt.IsZero() || now.Before(n.expiresAt)) {
			eligible = append(eligible, n)
		}
	}

	if len(eligible) == 0 {
		return ""
	}

	// 2. Weighted Selection (Simple Weighted Round Robin approach)
	// We sum weights and pick based on a running index
	totalWeight := 0
	for _, n := range eligible {
		w := n.weight
		if w <= 0 {
			w = 100 // default
		}
		totalWeight += w
	}

	// Pick a value in [0, totalWeight)
	currentIdx := atomic.AddUint32(&p.index, 1) % uint32(totalWeight)
	
	runningSum := 0
	for _, n := range eligible {
		w := n.weight
		if w <= 0 {
			w = 100
		}
		runningSum += w
		if uint32(runningSum) > currentIdx {
			return n.addr
		}
	}

	return eligible[0].addr // fallback
}

func registerDiscoveryAPI(r chi.Router) {
	r.Post("/api/zeno/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Service string `json:"service"`
			Host    string `json:"host"`
			Port    int    `json:"port"`
			Weight  int    `json:"weight"`
			TTL     int    `json:"ttl"` // seconds
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if req.Service == "" || req.Host == "" {
			http.Error(w, "Service and Host are required", http.StatusBadRequest)
			return
		}

		addr := req.Host
		if req.Port > 0 {
			addr = fmt.Sprintf("%s:%d", req.Host, req.Port)
		}

		registryMu.Lock()
		pool, exists := registry[req.Service]
		if !exists {
			pool = &servicePool{checkId: "/health"}
			registry[req.Service] = pool
		}
		registryMu.Unlock()

		pool.mu.Lock()
		defer pool.mu.Unlock()

		// Update or Add node
		found := false
		expiresAt := time.Time{}
		if req.TTL > 0 {
			expiresAt = time.Now().Add(time.Duration(req.TTL) * time.Second)
		} else {
			// If it's a dynamic registration without TTL, set a safe default of 5 mins
			expiresAt = time.Now().Add(5 * time.Minute)
		}

		for _, n := range pool.nodes {
			if n.addr == addr {
				n.weight = req.Weight
				n.expiresAt = expiresAt
				n.isHealthy = true // assume healthy on registration
				found = true
				break
			}
		}

		if !found {
			pool.nodes = append(pool.nodes, &serviceNode{
				addr:      addr,
				isHealthy: true,
				weight:    req.Weight,
				expiresAt: expiresAt,
			})
		}

		slog.Info("🔌 Dynamic Node Registered", "service", req.Service, "addr", addr, "weight", req.Weight)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "registered",
			"node":   addr,
			"expiry": expiresAt.Format(time.RFC3339),
		})
	})
}

func startServiceHealthChecker() {
	ticker := time.NewTicker(10 * time.Second)
	client := &http.Client{Timeout: 3 * time.Second}

	for range ticker.C {
		registryMu.RLock()
		pools := make(map[string]*servicePool)
		for k, v := range registry {
			pools[k] = v
		}
		registryMu.RUnlock()

		now := time.Now()
		for name, pool := range pools {
			pool.mu.Lock()
			
			// 1. [NEW] Prune Expired Nodes
			var activeNodes []*serviceNode
			for _, n := range pool.nodes {
				if n.expiresAt.IsZero() || now.Before(n.expiresAt) {
					activeNodes = append(activeNodes, n)
				} else {
					slog.Warn("⏳ Node Expired (Gracefully Removed)", "service", name, "addr", n.addr)
				}
			}
			pool.nodes = activeNodes

			// 2. Probing
			checkPath := pool.checkId
			if !strings.HasPrefix(checkPath, "/") {
				checkPath = "/" + checkPath
			}

			for _, node := range pool.nodes {
				url := fmt.Sprintf("http://%s%s", node.addr, checkPath)
				resp, err := client.Head(url)
				
				oldStatus := node.isHealthy
				node.isHealthy = (err == nil && resp.StatusCode < 500)
				node.lastCheck = time.Now()

				if oldStatus != node.isHealthy {
					if node.isHealthy {
						slog.Info("✅ Node is BACK ONLINE", "service", name, "addr", node.addr)
					} else {
						slog.Warn("❌ Node is DOWN", "service", name, "addr", node.addr, "error", err)
					}
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
			pool.mu.Unlock()
		}
	}
}
