package host

import (
	"sync"
)

type Manager struct {
	domains []string
	mu      sync.RWMutex
}

var GlobalManager = &Manager{}

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

func (m *Manager) GetDomains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]string, len(m.domains))
	copy(domains, m.domains)
	return domains
}
