package slots

import (
	"context"
	"testing"
	"time"
	"zeno/pkg/engine"
)

func TestTimeSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterTimeSlots(eng)

	t.Run("date.now", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "date.now",
			Children: []*engine.Node{
				{Name: "as", Value: "$current"},
				{Name: "format", Value: "yyyy-MM-dd"},
			},
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("date.now failed: %v", err)
		}
		val, _ := scope.Get("current")
		expected := time.Now().Format("2006-01-02")
		if val != expected {
			t.Errorf("Expected %s, got %v", expected, val)
		}
	})

	t.Run("date.format", func(t *testing.T) {
		scope := engine.NewScope(nil)
		testDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		scope.Set("my_date", testDate)

		node := &engine.Node{
			Name:  "date.format",
			Value: "$my_date",
			Children: []*engine.Node{
				{Name: "format", Value: "dd/MM/yyyy"},
				{Name: "as", Value: "$formatted"},
			},
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("date.format failed: %v", err)
		}
		val, _ := scope.Get("formatted")
		if val != "01/01/2024" {
			t.Errorf("Expected 01/01/2024, got %v", val)
		}
	})

	t.Run("date.parse", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name:  "date.parse",
			Value: "2024-05-20",
			Children: []*engine.Node{
				{Name: "as", Value: "$parsed"},
			},
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("date.parse failed: %v", err)
		}
		val, _ := scope.Get("parsed")
		parsedTime := val.(time.Time)
		if parsedTime.Year() != 2024 || parsedTime.Month() != 5 || parsedTime.Day() != 20 {
			t.Errorf("Incorrect parsed date: %v", parsedTime)
		}
	})

	t.Run("date.add", func(t *testing.T) {
		scope := engine.NewScope(nil)
		baseDate := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		scope.Set("start", baseDate)

		node := &engine.Node{
			Name:  "date.add",
			Value: "$start",
			Children: []*engine.Node{
				{Name: "duration", Value: "24h"},
				{Name: "as", Value: "$next_day"},
			},
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("date.add failed: %v", err)
		}
		val, _ := scope.Get("next_day")
		resultTime := val.(time.Time)
		if resultTime.Day() != 2 {
			t.Errorf("Expected day 2, got %v", resultTime.Day())
		}
	})
}
