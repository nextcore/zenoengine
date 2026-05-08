package slots

import (
	"reflect"
	"testing"
)

func TestGetNestedValue(t *testing.T) {
	data := map[string]interface{}{
		"title": "Report",
		"items": []interface{}{
			map[string]interface{}{"name": "A", "price": 100},
			map[string]interface{}{"name": "B", "price": 200},
		},
		"nested": map[string]interface{}{
			"deep": "value",
		},
	}

	tests := []struct {
		key  string
		want interface{}
	}{
		{"title", "Report"},
		{"items", data["items"]},
		{"nested.deep", "value"},
		{"items.name", []interface{}{"A", "B"}},
		{"items.price", []interface{}{100, 200}},
		{"items.missing", []interface{}{nil, nil}},
		{"invalid.key", nil},
	}

	for _, tt := range tests {
		got := getNestedValue(data, tt.key)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("getNestedValue(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}
