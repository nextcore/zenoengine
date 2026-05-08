package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/xuri/excelize/v2"
)

func RegisterExcelSlots(eng *engine.Engine) {
	eng.Register("excel.from_template", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return nil
		}

		// 1. Get Template Path
		templatePath := coerce.ToString(resolveValue(node.Value, scope))
		if templatePath == "" {
			return fmt.Errorf("excel.from_template: template path required")
		}

		// Clean path
		fullPath := templatePath
		if !filepath.IsAbs(fullPath) {
			cwd, _ := os.Getwd()
			fullPath = filepath.Join(cwd, templatePath)
		}
		slog.Info("Excel Export", "template", fullPath)

		// 2. Open File
		f, err := excelize.OpenFile(fullPath)
		if err != nil {
			slog.Error("Excel Open Failed", "error", err, "path", fullPath)
			return fmt.Errorf("excel.from_template: failed to open template '%s': %v", templatePath, err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		filename := "export.xlsx"
		sheetName := "Sheet1"
		markerData := make(map[string]interface{})

		// 3. Process Logic
		for _, c := range node.Children {
			// Metadata
			if c.Name == "filename" {
				filename = coerce.ToString(resolveValue(c.Value, scope))
			}
			if c.Name == "sheet" {
				sheetName = coerce.ToString(resolveValue(c.Value, scope))
			}

			// Single Cell Overrides
			if c.Name == "cell" {
				cellCoord := coerce.ToString(resolveValue(c.Value, scope))
				for _, child := range c.Children {
					if child.Name == "val" || child.Name == "value" {
						val := resolveValue(child.Value, scope)
						f.SetCellValue(sheetName, cellCoord, val)
					}
				}
			}

			// Batch Data & Markers
			if c.Name == "data" {
				dataVal := parseNodeValue(c, scope)
				if m, ok := dataVal.(map[string]interface{}); ok {
					for k, v := range m {
						// If key looks like a cell coordinate (A1, B2), set directly
						if isCellCoord(k) {
							f.SetCellValue(sheetName, k, v)
						} else {
							// Auto-detect JSON string in data
							// This handles cases where data comes as a JSON string from ZenoLang
							if strVal, ok := v.(string); ok {
								strVal = strings.TrimSpace(strVal)
								// Handle ZenoLang @json(...) directive passed as raw string
								// Regex: ^@\s*json\s*\((.*)\)$
								if strings.HasPrefix(strVal, "@") {
									// Simple manual parsing to avoid regex overhead/complexity
									lower := strings.ToLower(strVal)
									idxOpen := strings.Index(lower, "(")
									idxClose := strings.LastIndex(lower, ")")
									if idxOpen > 0 && idxClose > idxOpen && strings.Contains(lower, "json") {
										// Check if it's actually @...json...
										prefix := lower[:idxOpen]
										if strings.Contains(prefix, "json") {
											strVal = strVal[idxOpen+1 : idxClose]
											strVal = strings.TrimSpace(strVal)
										}
									}
								}

								// Strip outer quotes again if they were inside the @json(...)
								if len(strVal) >= 2 && ((strings.HasPrefix(strVal, "'") && strings.HasSuffix(strVal, "'")) ||
									(strings.HasPrefix(strVal, "\"") && strings.HasSuffix(strVal, "\""))) {
									strVal = strVal[1 : len(strVal)-1]
									strVal = strings.TrimSpace(strVal)
								}

								// Debug: print exactly what we represent
								// slog.Info("Checking JSON Candidate", "val_raw", strVal)
								if (strings.HasPrefix(strVal, "{") && strings.HasSuffix(strVal, "}")) ||
									(strings.HasPrefix(strVal, "[") && strings.HasSuffix(strVal, "]")) {
									var parsed interface{}
									// Try Unmarshal
									err := json.Unmarshal([]byte(strVal), &parsed)
									if err == nil {
										v = parsed
										slog.Info("JSON Auto-Parse Success", "key", k)
									} else {
										slog.Warn("JSON Auto-Parse Failed", "key", k, "error", err)
									}
								} else {
									slog.Info("JSON Candidate Rejected", "reason", "no_brace_bracket_prefix")
								}
							}

							// Otherwise treat as marker data
							markerData[k] = v
							slog.Info("Data Key", "key", k, "type", fmt.Sprintf("%T", v))
						}
					}
				}
			}
			// Formula Support
			if c.Name == "formulas" {
				val := parseNodeValue(c, scope)
				if m, ok := val.(map[string]interface{}); ok {
					for k, v := range m {
						formula := coerce.ToString(v)
						if isCellCoord(k) {
							if err := f.SetCellFormula(sheetName, k, formula); err != nil {
								slog.Warn("Excel SetFormula Failed", "cell", k, "formula", formula, "error", err)
							}
						}
					}
				}
			}

			// Image Support
			if c.Name == "images" {
				val := parseNodeValue(c, scope)
				// Expected: images: { "A1": "/path/to/image.png" }
				if m, ok := val.(map[string]interface{}); ok {
					for k, v := range m {
						imgPath := coerce.ToString(v)
						if !filepath.IsAbs(imgPath) {
							cwd, _ := os.Getwd()
							imgPath = filepath.Join(cwd, imgPath)
						}

						if isCellCoord(k) {
							// Basic image insertion
							// Scaling can be added via options if needed later
							err := f.AddPicture(sheetName, k, imgPath, &excelize.GraphicOptions{
								AutoFit: true,
							})
							if err != nil {
								slog.Warn("Excel AddPicture Failed", "cell", k, "path", imgPath, "error", err)
							}
						}
					}
				}
			}
		}

		// 4. Process Template Markers
		if len(markerData) > 0 {
			if err := processMarkers(f, sheetName, markerData); err != nil {
				slog.Error("Excel Marker Process Failed", "error", err)
				return err
			}
		}

		// 5. Send Response
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK) // Explicit Status OK

		slog.Info("Excel Writing Response...")
		if err := f.Write(w); err != nil {
			slog.Error("Excel Write Failed", "error", err)
			return fmt.Errorf("excel.from_template: failed to write response: %v", err)
		}
		slog.Info("Excel Export Complete")

		return nil
	}, engine.SlotMeta{
		Description: "Generate Excel from template with marker support",
		Example: `excel.from_template: 'template.xlsx'
  data:
    title: "Report"
    users: $user_list`,
	})
}

// Simple heuristic check for Cell Coordinate (e.g., A1, AA10, not "title")
func isCellCoord(s string) bool {
	matched, _ := regexp.MatchString(`^[A-Z]+[0-9]+$`, s)
	return matched
}

// ==========================================
// TEMPLATE MARKER LOGIC
// ==========================================

func processMarkers(f *excelize.File, sheet string, data map[string]interface{}) error {
	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`\{\{\s*([\w\.]+)\s*\}\}`)

	// Struct to track markers found in the sheet
	type Marker struct {
		Row, Col int
		Key      string
		Original string // {{ key }}
	}

	var markers []Marker

	// 1. Scan for all markers
	// Note: rows is [][]string. index i is row-1.
	for rIdx, row := range rows {
		for cIdx, cellValue := range row {
			matches := re.FindStringSubmatch(cellValue)
			if len(matches) > 1 {
				markers = append(markers, Marker{
					Row:      rIdx + 1, // 1-based index
					Col:      cIdx + 1, // 1-based index
					Key:      matches[1],
					Original: matches[0],
				})
			}
		}
	}

	// 2. Separate Scalar vs List Replacement
	// We need to process from BOTTOM to TOP for List expansions to avoid shifting issues affecting unprocessed markers.
	// Actually, sorting markers reversely by Row is sufficient.

	sort.Slice(markers, func(i, j int) bool {
		return markers[i].Row > markers[j].Row // Descending Sort
	})

	// Track inserted rows shift for debugging or future use?
	// Not needed if we process bottom-up.

	for _, m := range markers {
		// Get Value from Data
		val := getNestedValue(data, m.Key)

		// Check type
		if _, err := coerce.ToSlice(val); err == nil {
			// === LIST EXPANSION ===
			// Logic handled in Re-Process step below.
		} else {
			// === SCALAR REPLACEMENT ===
			cellName, _ := excelize.CoordinatesToCellName(m.Col, m.Row)
			f.SetCellValue(sheet, cellName, val)
		}
	}

	// 3. RE-PROCESS FOR LISTS (Row-Grouping)
	// We iterate original rows top-down? Or bottom-up?
	// Bottom-up is safer for insertion.

	// Group markers by Row
	rowMarkers := make(map[int][]Marker)
	for _, m := range markers {
		rowMarkers[m.Row] = append(rowMarkers[m.Row], m)
	}

	// Get sorted keys (Row Indices) Descending
	var rowsIndices []int
	for r := range rowMarkers {
		rowsIndices = append(rowsIndices, r)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(rowsIndices)))

	for _, rIdx := range rowsIndices {
		ms := rowMarkers[rIdx]

		// Check if any marker in this row is a List
		var listLen int = 0
		var isListRow bool = false

		for _, m := range ms {
			val := getNestedValue(data, m.Key)
			if slice, err := coerce.ToSlice(val); err == nil && len(slice) > 1 {
				if isListRow && len(slice) != listLen {
					// Mismatch lengths! Warning.
				}
				listLen = len(slice)
				isListRow = true
			}
		}

		if isListRow {
			// Insert rows
			// We need to insert N-1 rows
			numToInsert := listLen - 1
			if numToInsert > 0 {
				// Use InsertRows (plural) which takes count argument in v2
				if err := f.InsertRows(sheet, rIdx+1, numToInsert); err != nil {
					fmt.Printf("Error inserting rows: %v\n", err)
				}
			}

			// Prepare to fill data
			// For each marker in this row
			for _, m := range ms {
				val := getNestedValue(data, m.Key)
				slice, err := coerce.ToSlice(val)
				if err != nil {
					// Scalar in a list row?
					cellName, _ := excelize.CoordinatesToCellName(m.Col, rIdx)
					f.SetCellValue(sheet, cellName, val)
					continue
				}

				// Fill slice values
				for i := 0; i < listLen; i++ {
					if i >= len(slice) {
						break
					}
					// Write to Row rIdx + i
					cellName, _ := excelize.CoordinatesToCellName(m.Col, rIdx+i)
					f.SetCellValue(sheet, cellName, slice[i])
				}
			}

		} else {
			// Scalar Row - Just process values
			for _, m := range ms {
				val := getNestedValue(data, m.Key)
				originalVal, _ := f.GetCellValue(sheet, fmt.Sprintf("%s%d", getColName(m.Col), m.Row)) // simplified
				newVal := strings.Replace(originalVal, m.Original, coerce.ToString(val), -1)

				cellName, _ := excelize.CoordinatesToCellName(m.Col, m.Row)
				f.SetCellValue(sheet, cellName, newVal)
			}
		}
	}

	return nil
}

// Helper to get nested value "users.0.name"
func getNestedValue(data map[string]interface{}, key string) interface{} {
	parts := strings.Split(key, ".")
	var current interface{} = data

	for _, part := range parts {
		slog.Info("Nested Traverse", "part", part, "currentType", fmt.Sprintf("%T", current))
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil
			}
		} else if slice, ok := current.([]interface{}); ok {
			// List Projection: items.name -> [A, B, C]
			// We map the next part over the slice
			var projected []interface{}
			for _, item := range slice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if val, exists := itemMap[part]; exists {
						projected = append(projected, val)
					} else {
						projected = append(projected, nil)
					}
				}
			}
			current = projected
		} else {
			return nil
		}
	}
	return current
}

func getColName(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
