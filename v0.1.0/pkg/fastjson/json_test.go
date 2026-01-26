package fastjson

import (
	"testing"
	"encoding/json"
	"bytes"
)

type TestData struct {
	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Active  bool                   `json:"active"`
	Meta    map[string]interface{} `json:"meta"`
	Tags    []string               `json:"tags"`
}

func getTestData() *TestData {
	return &TestData{
		ID:     123,
		Name:   "Test User",
		Email:  "test@example.com",
		Active: true,
		Meta: map[string]interface{}{
			"role":  "admin",
			"score": 95.5,
		},
		Tags: []string{"golang", "performance", "optimization"},
	}
}

// Benchmark standard encoding/json
func BenchmarkStandardJSON(b *testing.B) {
	data := getTestData()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark goccy/go-json
func BenchmarkFastJSON(b *testing.B) {
	data := getTestData()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark encoder - standard
func BenchmarkStandardJSONEncoder(b *testing.B) {
	data := getTestData()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode(data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark encoder - fast
func BenchmarkFastJSONEncoder(b *testing.B) {
	data := getTestData()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := NewEncoder(buf)
		if err := enc.Encode(data); err != nil {
			b.Fatal(err)
		}
	}
}

// Test correctness
func TestJSONCompatibility(t *testing.T) {
	data := getTestData()

	// Marshal with both
	stdJSON, err1 := json.Marshal(data)
	fastJSON, err2 := Marshal(data)

	if err1 != nil || err2 != nil {
		t.Fatalf("Marshal errors: std=%v, fast=%v", err1, err2)
	}

	// Unmarshal both
	var stdResult, fastResult TestData
	if err := json.Unmarshal(stdJSON, &stdResult); err != nil {
		t.Fatal(err)
	}
	if err := Unmarshal(fastJSON, &fastResult); err != nil {
		t.Fatal(err)
	}

	// Compare
	if stdResult.ID != fastResult.ID {
		t.Error("ID mismatch")
	}
	if stdResult.Name != fastResult.Name {
		t.Error("Name mismatch")
	}
	if stdResult.Email != fastResult.Email {
		t.Error("Email mismatch")
	}
}
