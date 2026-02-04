package slots

import (
	"context"
	"fmt"
	"strings"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Exported for testing
var GetSheetsServiceFunc = func(ctx context.Context, creds interface{}) (*sheets.Service, error) {
	credsStr := coerce.ToString(creds)
	if credsStr == "" {
		return nil, fmt.Errorf("google sheets: credentials required (JSON string or path to file)")
	}

	var opt option.ClientOption
	if strings.HasPrefix(strings.TrimSpace(credsStr), "{") {
		opt = option.WithCredentialsJSON([]byte(credsStr))
	} else {
		opt = option.WithCredentialsFile(credsStr)
	}

	return sheets.NewService(ctx, opt)
}

func RegisterGSheetSlots(eng *engine.Engine) {
	// ==========================================
	// SLOT: GSHEET.GET
	// ==========================================
	eng.Register("gsheet.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, gRange, creds interface{}
		target := "rows"

		for _, c := range node.Children {
			switch c.Name {
			case "id", "spreadsheet_id":
				id = resolveValue(c.Value, scope)
			case "range":
				gRange = resolveValue(c.Value, scope)
			case "credentials", "creds":
				creds = resolveValue(c.Value, scope)
			case "as":
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if id == nil {
			id = resolveValue(node.Value, scope)
		}

		if id == nil || gRange == nil || creds == nil {
			return fmt.Errorf("gsheet.get: id, range, and credentials are required")
		}

		srv, err := GetSheetsServiceFunc(ctx, creds)
		if err != nil {
			return err
		}

		resp, err := srv.Spreadsheets.Values.Get(coerce.ToString(id), coerce.ToString(gRange)).Do()
		if err != nil {
			return fmt.Errorf("gsheet.get: %v", err)
		}

		scope.Set(target, resp.Values)
		return nil
	}, engine.SlotMeta{
		Description: "Fetch data from a Google Spreadsheet.",
		ValueType:   "string",
		Example: `gsheet.get: $spreadsheet_id
  range: "Sheet1!A1:B10"
  credentials: $service_account_json
  as: $rows`,
		Inputs: map[string]engine.InputMeta{
			"id":          {Description: "Spreadsheet ID", Required: true, Type: "string"},
			"range":       {Description: "Cell range (e.g. Sheet1!A1:B10)", Required: true, Type: "string"},
			"credentials": {Description: "Service account JSON or path to file", Required: true, Type: "string"},
			"as":          {Description: "Variable to store results (Default: rows)", Required: false, Type: "string"},
		},
	})

	// ==========================================
	// SLOT: GSHEET.APPEND
	// ==========================================
	eng.Register("gsheet.append", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, gRange, creds, values interface{}

		for _, c := range node.Children {
			switch c.Name {
			case "id", "spreadsheet_id":
				id = resolveValue(c.Value, scope)
			case "range":
				gRange = resolveValue(c.Value, scope)
			case "credentials", "creds":
				creds = resolveValue(c.Value, scope)
			case "values", "val":
				values = resolveValue(c.Value, scope)
			}
		}

		if id == nil {
			id = resolveValue(node.Value, scope)
		}

		if id == nil || gRange == nil || creds == nil || values == nil {
			return fmt.Errorf("gsheet.append: id, range, values, and credentials are required")
		}

		srv, err := GetSheetsServiceFunc(ctx, creds)
		if err != nil {
			return err
		}

		sliceValues, err := coerce.ToSlice(values)
		if err != nil {
			return fmt.Errorf("gsheet.append: values must be a list of lists")
		}

		var rowData [][]interface{}
		for _, row := range sliceValues {
			rowSlice, err := coerce.ToSlice(row)
			if err != nil {
				// If it's not a slice, treat it as a single-column row
				rowData = append(rowData, []interface{}{row})
			} else {
				rowData = append(rowData, rowSlice)
			}
		}

		rb := &sheets.ValueRange{
			Values: rowData,
		}

		_, err = srv.Spreadsheets.Values.Append(coerce.ToString(id), coerce.ToString(gRange), rb).ValueInputOption("RAW").Do()
		if err != nil {
			return fmt.Errorf("gsheet.append: %v", err)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Append rows to a Google Spreadsheet.",
		ValueType:   "string",
		Example: `gsheet.append: $spreadsheet_id
  range: "Sheet1!A1"
  values: $new_rows
  credentials: $service_account_json`,
		Inputs: map[string]engine.InputMeta{
			"id":          {Description: "Spreadsheet ID", Required: true, Type: "string"},
			"range":       {Description: "Range to find the end of table (e.g. Sheet1!A1)", Required: true, Type: "string"},
			"values":      {Description: "List of rows to append", Required: true, Type: "list"},
			"credentials": {Description: "Service account JSON or path to file", Required: true, Type: "string"},
		},
	})

	// ==========================================
	// SLOT: GSHEET.UPDATE
	// ==========================================
	eng.Register("gsheet.update", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, gRange, creds, values interface{}

		for _, c := range node.Children {
			switch c.Name {
			case "id", "spreadsheet_id":
				id = resolveValue(c.Value, scope)
			case "range":
				gRange = resolveValue(c.Value, scope)
			case "credentials", "creds":
				creds = resolveValue(c.Value, scope)
			case "values", "val":
				values = resolveValue(c.Value, scope)
			}
		}

		if id == nil {
			id = resolveValue(node.Value, scope)
		}

		if id == nil || gRange == nil || creds == nil || values == nil {
			return fmt.Errorf("gsheet.update: id, range, values, and credentials are required")
		}

		srv, err := GetSheetsServiceFunc(ctx, creds)
		if err != nil {
			return err
		}

		sliceValues, err := coerce.ToSlice(values)
		if err != nil {
			return fmt.Errorf("gsheet.update: values must be a list of lists")
		}

		var rowData [][]interface{}
		for _, row := range sliceValues {
			rowSlice, err := coerce.ToSlice(row)
			if err != nil {
				rowData = append(rowData, []interface{}{row})
			} else {
				rowData = append(rowData, rowSlice)
			}
		}

		rb := &sheets.ValueRange{
			Values: rowData,
		}

		_, err = srv.Spreadsheets.Values.Update(coerce.ToString(id), coerce.ToString(gRange), rb).ValueInputOption("RAW").Do()
		if err != nil {
			return fmt.Errorf("gsheet.update: %v", err)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Update a specific range in a Google Spreadsheet.",
		ValueType:   "string",
		Example: `gsheet.update: $spreadsheet_id
  range: "Sheet1!A1:B2"
  values: [['Name', 'Age'], ['Budi', 25]]
  credentials: $service_account_json`,
		Inputs: map[string]engine.InputMeta{
			"id":          {Description: "Spreadsheet ID", Required: true, Type: "string"},
			"range":       {Description: "Range to update", Required: true, Type: "string"},
			"values":      {Description: "New data to write", Required: true, Type: "list"},
			"credentials": {Description: "Service account JSON or path to file", Required: true, Type: "string"},
		},
	})

	// ==========================================
	// SLOT: GSHEET.FIND
	// ==========================================
	eng.Register("gsheet.find", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, gRange, creds, where interface{}
		target := "results"

		for _, c := range node.Children {
			switch c.Name {
			case "id", "spreadsheet_id":
				id = resolveValue(c.Value, scope)
			case "range":
				gRange = resolveValue(c.Value, scope)
			case "credentials", "creds":
				creds = resolveValue(c.Value, scope)
			case "where":
				where = resolveValue(c.Value, scope)
			case "as":
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if id == nil {
			id = resolveValue(node.Value, scope)
		}

		if id == nil || gRange == nil || creds == nil || where == nil {
			return fmt.Errorf("gsheet.find: id, range, credentials, and where are required")
		}

		srv, err := GetSheetsServiceFunc(ctx, creds)
		if err != nil {
			return err
		}

		resp, err := srv.Spreadsheets.Values.Get(coerce.ToString(id), coerce.ToString(gRange)).Do()
		if err != nil {
			return fmt.Errorf("gsheet.find: %v", err)
		}

		if len(resp.Values) < 1 {
			scope.Set(target, []interface{}{})
			return nil
		}

		// 1. Process Headers
		headers := make([]string, len(resp.Values[0]))
		for i, h := range resp.Values[0] {
			headers[i] = coerce.ToString(h)
		}

		// 2. Process Criteria
		criteria, ok := where.(map[string]interface{})
		if !ok {
			return fmt.Errorf("gsheet.find: 'where' must be a map of { Column: Value }")
		}

		results := []interface{}{}

		// 3. Search Loop
		for rIdx := 1; rIdx < len(resp.Values); rIdx++ {
			row := resp.Values[rIdx]
			rowMap := make(map[string]interface{})

			// Build map for this row
			for i, h := range headers {
				if i < len(row) {
					rowMap[h] = row[i]
				} else {
					rowMap[h] = nil
				}
			}

			// Add metadata
			rowMap["__row"] = rIdx + 1

			// Check Match
			match := true
			for k, v := range criteria {
				rowVal := rowMap[k]
				if coerce.ToString(rowVal) != coerce.ToString(v) {
					match = false
					break
				}
			}

			if match {
				results = append(results, rowMap)
			}
		}

		scope.Set(target, results)
		return nil
	}, engine.SlotMeta{
		Description: "Search for rows in a Google Spreadsheet using header-based filtering.",
		ValueType:   "string",
		Example: `gsheet.find: $spreadsheet_id
  range: "Sheet1!A1:Z100"
  where: { "Nama": "Budi", "Status": "Active" }
  credentials: $service_account_json
  as: $rows`,
		Inputs: map[string]engine.InputMeta{
			"id":          {Description: "Spreadsheet ID", Required: true, Type: "string"},
			"range":       {Description: "Cell range including headers (e.g. Sheet1!A1:Z100)", Required: true, Type: "string"},
			"where":       {Description: "Criteria map (column header names as keys)", Required: true, Type: "map"},
			"credentials": {Description: "Service account JSON or path to file", Required: true, Type: "string"},
			"as":          {Description: "Variable to store results (Default: results)", Required: false, Type: "string"},
		},
	})

	// ==========================================
	// SLOT: GSHEET.CLEAR
	// ==========================================
	eng.Register("gsheet.clear", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, gRange, creds interface{}

		for _, c := range node.Children {
			switch c.Name {
			case "id", "spreadsheet_id":
				id = resolveValue(c.Value, scope)
			case "range":
				gRange = resolveValue(c.Value, scope)
			case "credentials", "creds":
				creds = resolveValue(c.Value, scope)
			}
		}

		if id == nil {
			id = resolveValue(node.Value, scope)
		}

		if id == nil || gRange == nil || creds == nil {
			return fmt.Errorf("gsheet.clear: id, range, and credentials are required")
		}

		srv, err := GetSheetsServiceFunc(ctx, creds)
		if err != nil {
			return err
		}

		_, err = srv.Spreadsheets.Values.Clear(coerce.ToString(id), coerce.ToString(gRange), &sheets.ClearValuesRequest{}).Do()
		if err != nil {
			return fmt.Errorf("gsheet.clear: %v", err)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Clear data from a specific range in a Google Spreadsheet.",
		ValueType:   "string",
		Example: `gsheet.clear: $spreadsheet_id
  range: "Sheet1!A1:B10"
  credentials: $service_account_json`,
		Inputs: map[string]engine.InputMeta{
			"id":          {Description: "Spreadsheet ID", Required: true, Type: "string"},
			"range":       {Description: "Range to clear", Required: true, Type: "string"},
			"credentials": {Description: "Service account JSON or path to file", Required: true, Type: "string"},
		},
	})
}
