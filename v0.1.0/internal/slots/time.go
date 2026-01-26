package slots

import (
	"context"
	"fmt"
	"strings"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterTimeSlots(eng *engine.Engine) {
	// 1. DATE.NOW
	eng.Register("date.now", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		now := time.Now()
		target := "now"
		layout := time.RFC3339

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "layout" || c.Name == "format" {
				layout = resolveLayout(coerce.ToString(parseNodeValue(c, scope)))
			}
		}

		// Support simple string output by default or object if requested
		scope.Set(target, now.Format(layout))
		scope.Set(target+"_obj", now)
		return nil
	}, engine.SlotMeta{
		Description: "Mengambil waktu saat ini.",
		Example:     "date.now: { as: $skarang }",
		Inputs: map[string]engine.InputMeta{
			"as":     {Description: "Variabel penyimpan hasil string"},
			"layout": {Description: "Format tanggal (RFC3339, Human, dll)"},
			"format": {Description: "Alias untuk layout"},
		},
	})

	// 2. DATE.FORMAT
	eng.Register("date.format", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var input interface{}
		layout := "2006-01-02 15:04:05"
		target := "formatted_date"

		if node.Value != nil {
			val, err := MustResolveValue(node.Value, scope, coerce.ToString(node.Value))
			if err != nil {
				return err
			}
			input = val
		}

		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "date" {
				input = parseNodeValue(c, scope)
			}
			if c.Name == "layout" || c.Name == "format" {
				layout = resolveLayout(coerce.ToString(parseNodeValue(c, scope)))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if input == nil {
			return fmt.Errorf("date.format: input date is required")
		}

		var t time.Time
		switch v := input.(type) {
		case time.Time:
			t = v
		case string:
			parsed, err := parseFlexDate(v)
			if err != nil {
				return fmt.Errorf("date.format: failed to parse input string '%s': %v", v, err)
			}
			t = parsed
		default:
			return fmt.Errorf("date.format: invalid input type %T", input)
		}

		scope.Set(target, t.Format(layout))
		return nil
	}, engine.SlotMeta{
		Description: "Memformat objek tanggal atau string tanggal.",
		Example:     "date.format: $created_at { layout: 'Human'; as: $tgl }",
		Inputs: map[string]engine.InputMeta{
			"val":    {Description: "Objek atau string tanggal"},
			"date":   {Description: "Alias untuk val"},
			"layout": {Description: "Format tujuan"},
			"format": {Description: "Alias untuk layout"},
			"as":     {Description: "Variabel penyimpan hasil"},
		},
	})

	// 3. DATE.PARSE
	eng.Register("date.parse", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		input := coerce.ToString(resolveValue(node.Value, scope))
		layout := ""
		target := "parsed_date"

		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "input" {
				input = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "layout" || c.Name == "format" {
				layout = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if input == "" {
			return fmt.Errorf("date.parse: input string is required")
		}

		var t time.Time
		var err error
		if layout != "" {
			t, err = time.Parse(resolveLayout(layout), input)
		} else {
			t, err = parseFlexDate(input)
		}

		if err != nil {
			return fmt.Errorf("date.parse: %v", err)
		}

		scope.Set(target, t)
		return nil
	}, engine.SlotMeta{
		Description: "Mengubah string menjadi objek tanggal.",
		Example:     "date.parse: '2023-12-25' { as: $tgl_obj }",
		Inputs: map[string]engine.InputMeta{
			"input":  {Description: "String tanggal"},
			"val":    {Description: "Alias untuk input"},
			"layout": {Description: "Format sumber"},
			"format": {Description: "Alias untuk layout"},
			"as":     {Description: "Variabel penyimpan hasil"},
		},
	})

	// 4. DATE.ADD
	eng.Register("date.add", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var input interface{}
		durationStr := ""
		target := "shifted_date"

		if node.Value != nil {
			input = resolveValue(node.Value, scope)
		}

		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "date" {
				input = parseNodeValue(c, scope)
			}
			if c.Name == "duration" || c.Name == "add" {
				durationStr = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if input == nil {
			input = time.Now()
		}

		var t time.Time
		switch v := input.(type) {
		case time.Time:
			t = v
		case string:
			parsed, err := parseFlexDate(v)
			if err != nil {
				return fmt.Errorf("date.add: invalid date string '%s'", v)
			}
			t = parsed
		default:
			return fmt.Errorf("date.add: invalid input type")
		}

		d, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("date.add: invalid duration '%s' (use 1h, 30m, etc)", durationStr)
		}

		scope.Set(target, t.Add(d))
		return nil
	}, engine.SlotMeta{
		Description: "Menambah durasi ke tanggal.",
		Example:     "date.add: $now { duration: '2h'; as: $future }",
		Inputs: map[string]engine.InputMeta{
			"date":     {Description: "Objek tanggal sumber"},
			"val":      {Description: "Alias untuk date"},
			"duration": {Description: "Durasi (1h, 30m, 10s)"},
			"add":      {Description: "Alias untuk duration"},
			"as":       {Description: "Variabel penyimpan hasil"},
		},
	})
}

// Helper: Resolve human aliases to Go layouts
func resolveLayout(layout string) string {
	switch strings.ToLower(layout) {
	case "human":
		return "02 Jan 2006 15:04"
	case "date":
		return "2006-01-02"
	case "time":
		return "15:04:05"
	case "rfc3339":
		return time.RFC3339
	case "full":
		return "Monday, 02 January 2006 15:04:05"
	default:
		// Jika mengandung token seperti yyyy, MM, dd, anggap format custom ala C#/PHP
		if strings.ContainsAny(layout, "yMdHms") {
			return translateCustomLayout(layout)
		}
		return layout
	}
}

// translateCustomLayout mengonversi token umum (yyyy, MM, dd, HH, mm, ss) ke layout Go (2006, 01, 02, 15, 04, 05)
func translateCustomLayout(layout string) string {
	replacer := strings.NewReplacer(
		"yyyy", "2006",
		"yy", "06",
		"MMMM", "January",
		"MMM", "Jan",
		"MM", "01",
		"M", "1",
		"dddd", "Monday",
		"ddd", "Mon",
		"dd", "02",
		"d", "2",
		"HH", "15",
		"hh", "03",
		"h", "3",
		"mm", "04",
		"m", "4",
		"ss", "05",
		"s", "5",
		"tt", "PM",
	)

	return replacer.Replace(layout)
}

// Helper: Try multiple common layouts
func parseFlexDate(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
		"02 Jan 2006",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format")
}
