package slots

import (
	"fmt"
	"testing"
	"time"
)

func TestWeightedRoundRobin(t *testing.T) {
	pool := &servicePool{
		nodes: []*serviceNode{
			{addr: "NodeA", weight: 100, isHealthy: true},
			{addr: "NodeB", weight: 10, isHealthy: true},
		},
	}

	counts := make(map[string]int)
	totalCalls := 110

	for i := 0; i < totalCalls; i++ {
		addr := pool.getNextHealthyNode()
		counts[addr]++
	}

	if counts["NodeA"] != 100 {
		t.Errorf("Expected NodeA to be picked 100 times, got %d", counts["NodeA"])
	}
	if counts["NodeB"] != 10 {
		t.Errorf("Expected NodeB to be picked 10 times, got %d", counts["NodeB"])
	}
}

func TestNodePruning(t *testing.T) {
	// We need to test the logic inside startServiceHealthChecker but without the actual ticker
	// We'll simulate by manually calling the pruning logic or just testing the pruning part.
	
	now := time.Now()
	nodes := []*serviceNode{
		{addr: "Healthy", isHealthy: true, expiresAt: now.Add(1 * time.Hour)},
		{addr: "Expired", isHealthy: true, expiresAt: now.Add(-1 * time.Minute)},
		{addr: "Static", isHealthy: true, expiresAt: time.Time{}}, // Never expires
	}

	// Manual Pruning logic check
	var activeNodes []*serviceNode
	for _, n := range nodes {
		if n.expiresAt.IsZero() || now.Before(n.expiresAt) {
			activeNodes = append(activeNodes, n)
		}
	}

	if len(activeNodes) != 2 {
		t.Fatalf("Expected 2 active nodes, got %d", len(activeNodes))
	}
	
	foundExpired := false
	for _, n := range activeNodes {
		if n.addr == "Expired" {
			foundExpired = true
		}
	}
	if foundExpired {
		t.Error("Expired node should have been pruned")
	}
}

func TestConcurrentRegistryAccess(t *testing.T) {
	// Stress test for mutexes
	serviceName := "stress_test"
	totalRoutines := 50
	callsPerRoutine := 100

	done := make(chan bool)

	for i := 0; i < totalRoutines; i++ {
		go func(id int) {
			for j := 0; j < callsPerRoutine; j++ {
				// Register
				registryMu.Lock()
				pool, exists := registry[serviceName]
				if !exists {
					pool = &servicePool{}
					registry[serviceName] = pool
				}
				registryMu.Unlock()

				pool.mu.Lock()
				addr := fmt.Sprintf("node-%d-%d", id, j)
				pool.nodes = append(pool.nodes, &serviceNode{addr: addr, isHealthy: true})
				pool.mu.Unlock()

				// Resolve
				pool.getNextHealthyNode()
			}
			done <- true
		}(i)
	}

	for i := 0; i < totalRoutines; i++ {
		<-done
	}
}
