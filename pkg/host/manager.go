package host

import (
	"net/http"
	"sync"
)

type Manager struct {
	domains []string
	routers map[string]http.Handler // Map for O(1) lookup
	mu      sync.RWMutex
}

var GlobalManager = &Manager{
	routers: make(map[string]http.Handler),
}

func (m *Manager) RegisterDomain(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, d := range m.domains {
		if d == domain {
			return
		}
	}
	m.domains = append(m.domains, domain)
}

func (m *Manager) RegisterRouter(domain string, handler http.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure domain is in the list for ACME
	found := false
	for _, d := range m.domains {
		if d == domain {
			found = true
			break
		}
	}
	if !found {
		m.domains = append(m.domains, domain)
	}

	m.routers[domain] = handler
}

func (m *Manager) GetHandler(domain string) (http.Handler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.routers[domain]
	return h, ok
}

func (m *Manager) GetDomains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]string, len(m.domains))
	copy(domains, m.domains)
	return domains
}

func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.domains = []string{}
	m.routers = make(map[string]http.Handler)
}
