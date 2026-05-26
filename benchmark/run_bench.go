package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	TotalRequests = 5000
	Concurrency   = 50
	Payload       = `{"script": "var: $a { val: 10 }\nvar: $b { val: 20 }\nmath.calc: \"$a + $b\" {\n  as: $sum\n}\ndate.now: {\n  layout: \"yyyy-MM-dd\"\n  as: $today\n}"}`
)

type BenchResult struct {
	Target       string
	RPS          float64
	AvgLatencyMs float64
	TotalTimeSec float64
	SuccessCount int
	ErrorCount   int
}

func benchmark(url string, name string) BenchResult {
	fmt.Printf("\n⚡ Benchmarking %s (%s)...\n", name, url)

	// Pre-build request body
	bodyBytes := []byte(Payload)

	// Keep-alive HTTP Client
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: Concurrency,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 5 * time.Second,
	}

	jobs := make(chan struct{}, TotalRequests)
	for i := 0; i < TotalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var wg sync.WaitGroup
	var mu sync.Mutex

	var totalLatency time.Duration
	successCount := 0
	errorCount := 0

	startTime := time.Now()

	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				reqStart := time.Now()
				req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}

				// Consume body to allow connection reuse
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				latency := time.Since(reqStart)

				mu.Lock()
				if resp.StatusCode == http.StatusOK {
					successCount++
					totalLatency += latency
				} else {
					errorCount++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	var avgLatency float64
	if successCount > 0 {
		avgLatency = float64(totalLatency.Milliseconds()) / float64(successCount)
	}

	rps := float64(successCount+errorCount) / totalTime.Seconds()

	return BenchResult{
		Target:       name,
		RPS:          rps,
		AvgLatencyMs: avgLatency,
		TotalTimeSec: totalTime.Seconds(),
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	}
}

func main() {
	fmt.Printf("==================================================\n")
	fmt.Printf("        ZENOENGINE CONCURRENT HTTP BENCHMARK      \n")
	fmt.Printf("==================================================\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Total Requests: %d\n", TotalRequests)
	fmt.Printf("  Concurrency:    %d\n", Concurrency)
	fmt.Printf("==================================================\n")

	// 1. Bench Go
	goResult := benchmark("http://127.0.0.1:4000/execute", "Go ZenoEngine (std net/http)")

	// 2. Bench Rust
	rustResult := benchmark("http://127.0.0.1:3000/execute", "Rust ZenoEngine (Axum)")

	// Print Summary
	fmt.Printf("\n==================================================\n")
	fmt.Printf("                    SUMMARY RESULTS               \n")
	fmt.Printf("==================================================\n")
	fmt.Printf("%-30s | %-12s | %-12s | %-12s\n", "Target", "RPS", "Avg Latency", "Success/Error")
	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("%-30s | %-12.2f | %-10.2fms | %d/%d\n",
		goResult.Target, goResult.RPS, goResult.AvgLatencyMs, goResult.SuccessCount, goResult.ErrorCount)
	fmt.Printf("%-30s | %-12.2f | %-10.2fms | %d/%d\n",
		rustResult.Target, rustResult.RPS, rustResult.AvgLatencyMs, rustResult.SuccessCount, rustResult.ErrorCount)
	fmt.Printf("==================================================\n")

	speedup := rustResult.RPS / goResult.RPS
	fmt.Printf("🚀 Rust is %.2fx faster than Go!\n", speedup)
	fmt.Printf("==================================================\n")
}
