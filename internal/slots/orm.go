package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterORMSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// ORM.MODEL: 'users'
	eng.Register("orm.model", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		tableName := coerce.ToString(resolveValue(node.Value, scope))
		dbName := "default"

		for _, c := range node.Children {
			if c.Name == "db" || c.Name == "connection" {
				dbName = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		// Leverage existing db.table logic by setting _query_state
		dialect := dbMgr.GetDialect(dbName)
		scope.Set("_query_state", &QueryState{
			Table:   tableName,
			DBName:  dbName,
			Dialect: dialect,
		})

		// Store model metadata for other orm.* slots
		scope.Set("_active_model", tableName)

		// Execute children to allow registering relationships inside the block
		// Parse relations and schema metadata directly from children instead of evaluating
		for _, c := range node.Children {
			if c.Name == "db" || c.Name == "connection" || c.Name == "table" || c.Name == "name" {
				continue
			}

			// Capture Schema Properties directly
			if c.Name == "fillable" {
				fillableStr := coerce.ToString(parseNodeValue(c, scope))
				fillableParts := strings.Split(fillableStr, ",")

				fillMap := make(map[string]bool)
				for _, f := range fillableParts {
					fillMap[strings.TrimSpace(f)] = true
				}
				scope.Set("_schema_"+tableName+"_fillable", fillMap)
				continue
			}

			// Execute relationship decorators inside the model block
			// Because standard execution of unknown nodes inside 'orm.model' can cause infinite loops
			// we explicitly map only known relation keywords
			if strings.HasPrefix(c.Name, "orm.hasMany") || strings.HasPrefix(c.Name, "orm.hasOne") || strings.HasPrefix(c.Name, "orm.belongsTo") || strings.HasPrefix(c.Name, "orm.belongsToMany") {
				if err := eng.Execute(ctx, c, scope); err != nil {
					return err
				}
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Define the active model/table for ORM operations.",
		Example:     "orm.model: 'users'",
	})

	// ORM.FIND: 1 { as: $user }
	eng.Register("orm.find", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		id := resolveValue(node.Value, scope)
		target := "model"
		primaryKey := "id"

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "key" || c.Name == "pk" {
				primaryKey = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		// Ensure query state exists
		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("orm.find: no model defined. Call orm.model first")
		}
		qs := qsVal.(*QueryState)

		// Create a temporary filter for find
		originalWhere := qs.Where
		qs.Where = append(qs.Where, WhereCond{Column: primaryKey, Op: "=", Value: id})

		// Use db.first logic (Execute db.first slot internally)
		firstNode := &engine.Node{
			Name: "db.first",
			Children: []*engine.Node{
				{Name: "as", Value: target},
			},
		}

		err := eng.Execute(ctx, firstNode, scope)

		// Restore original where state
		qs.Where = originalWhere

		return err
	}, engine.SlotMeta{
		Description: "Find a single record by primary key.",
		Example:     "orm.find: 1 { as: $user }",
	})

	// ORM.SAVE: $user
	eng.Register("orm.save", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		dataRaw := resolveValue(node.Value, scope)
		data, ok := dataRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("orm.save: expected map data, got %T", dataRaw)
		}

		primaryKey := "id"
		for _, c := range node.Children {
			if c.Name == "key" || c.Name == "pk" {
				primaryKey = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("orm.save: no model defined")
		}
		qs := qsVal.(*QueryState)

		idVal, hasId := data[primaryKey]
		// Check if ID exists and is non-zero
		isUpdate := false
		if hasId && idVal != nil {
			idInt, _ := coerce.ToInt(idVal)
			if idInt > 0 {
				isUpdate = true
			}
		}

		// Apply Fillable Filter if defined
		isFillable := func(key string) bool {
			if key == primaryKey {
				// We don't want to update primary keys usually, but for insert it shouldn't matter as it falls through
				// However, if we're filtering mass assignment, we shouldn't allow manually updating the primary key.
				return false
			}

			fillableRaw, ok := scope.Get("_schema_" + qs.Table + "_fillable")
			if !ok || fillableRaw == nil {
				return true // No fillable defined, allow all
			}
			fillMap := fillableRaw.(map[string]bool)
			return fillMap[key]
		}
		sanitizeValue := func(v interface{}) interface{} {
			switch val := v.(type) {
			case map[string]interface{}, []interface{}, []map[string]interface{}:
				b, err := json.Marshal(val)
				if err == nil {
					return string(b)
				}
			}
			return v
		}

		executor, dialect, err := getExecutor(scope, dbMgr, qs.DBName)
		if err != nil {
			return err
		}

		if isUpdate {
			// Update Logic
			var sets []string
			var vals []interface{}

			i := 1
			for k, v := range data {
				if !isFillable(k) {
					continue
				}
				sets = append(sets, fmt.Sprintf("%s = %s", dialect.QuoteIdentifier(k), dialect.Placeholder(i)))
				vals = append(vals, sanitizeValue(v))
				i++
			}

			if len(sets) == 0 {
				return nil // Nothing to update
			}

			// Add primary key to where clause
			whereClause := ""
			baseIdx := len(vals)

			// We only want the PK for the where condition of the update
			whereConds := []string{
				fmt.Sprintf("%s = %s", dialect.QuoteIdentifier(primaryKey), dialect.Placeholder(baseIdx+1)),
			}
			vals = append(vals, idVal)
			whereClause = " WHERE " + strings.Join(whereConds, " AND ")

			query := fmt.Sprintf("UPDATE %s SET %s%s", dialect.QuoteIdentifier(qs.Table), strings.Join(sets, ", "), whereClause)
			fmt.Printf("Executing update: %s\n", query)
			_, err = executor.ExecContext(ctx, query, vals...)
			return err

		} else {
			// Insert Logic
			var cols []string
			var placeholders []string
			var vals []interface{}

			i := 1
			for k, v := range data {
				// Don't insert PK if it's nil/0
				if k == primaryKey {
					continue
				}
				if !isFillable(k) {
					continue
				}
				cols = append(cols, dialect.QuoteIdentifier(k))
				placeholders = append(placeholders, dialect.Placeholder(i))
				vals = append(vals, sanitizeValue(v))
				i++
			}

			if len(cols) == 0 {
				return nil // Nothing to insert
			}

			query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				dialect.QuoteIdentifier(qs.Table), strings.Join(cols, ", "), strings.Join(placeholders, ", "))
			fmt.Printf("Executing insert: %s\n", query)

			res, err := executor.ExecContext(ctx, query, vals...)
			if err == nil {
				if dialect.Name() != "postgres" {
					if lastId, err := res.LastInsertId(); err == nil {
						data[primaryKey] = lastId
						scope.Set("db_last_id", lastId)
					}
				}
			}
			return err
		}
	}, engine.SlotMeta{
		Description: "Save (Insert or Update) a model object.",
		Example:     "orm.save: $user",
	})

	// ORM.DELETE: $user (or ID)
	eng.Register("orm.delete", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		val := resolveValue(node.Value, scope)
		primaryKey := "id"
		var id interface{}

		if data, ok := val.(map[string]interface{}); ok {
			id = data[primaryKey]
		} else {
			id = val
		}

		if id == nil {
			return fmt.Errorf("orm.delete: ID not found")
		}

		qsVal, ok := scope.Get("_query_state")
		if !ok {
			return fmt.Errorf("orm.delete: no model defined")
		}
		qs := qsVal.(*QueryState)

		originalWhere := qs.Where
		qs.Where = append(qs.Where, WhereCond{Column: primaryKey, Op: "=", Value: id})

		deleteNode := &engine.Node{Name: "db.delete"}
		err := eng.Execute(ctx, deleteNode, scope)

		qs.Where = originalWhere
		return err
	}, engine.SlotMeta{})

	// ORM.BELONGSTO: 'User' { as: 'author', foreign_key: 'user_id' }
	eng.Register("orm.belongsTo", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		relatedModel := coerce.ToString(resolveValue(node.Value, scope))
		asName := strings.ToLower(relatedModel)
		foreignKey := asName + "_id"

		for _, c := range node.Children {
			if c.Name == "as" {
				asName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "foreign_key" || c.Name == "fk" {
				foreignKey = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		// Store relationship info in current model's metadata
		modelName := coerce.ToString(scope.GetDefault("_active_model", ""))
		if modelName == "" {
			return fmt.Errorf("orm.belongsTo: no active model")
		}

		relKey := fmt.Sprintf("_rel_%s_%s", modelName, asName)
		scope.Set(relKey, map[string]interface{}{
			"type":        "belongsTo",
			"model":       relatedModel,
			"foreign_key": foreignKey,
		})

		return nil
	}, engine.SlotMeta{Description: "Define a many-to-one relationship."})

	// ORM.HASMANY: 'Post' { as: 'posts', foreign_key: 'user_id' }
	eng.Register("orm.hasMany", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		relatedModel := coerce.ToString(resolveValue(node.Value, scope))
		asName := strings.ToLower(relatedModel) + "s"
		localKey := "id"
		foreignKey := strings.ToLower(coerce.ToString(scope.GetDefault("_active_model", ""))) + "_id"

		for _, c := range node.Children {
			if c.Name == "as" {
				asName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "foreign_key" || c.Name == "fk" {
				foreignKey = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "local_key" || c.Name == "lk" {
				localKey = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		modelName := coerce.ToString(scope.GetDefault("_active_model", ""))
		if modelName == "" {
			return fmt.Errorf("orm.hasMany: no active model")
		}

		relKey := fmt.Sprintf("_rel_%s_%s", modelName, asName)
		scope.Set(relKey, map[string]interface{}{
			"type":        "hasMany",
			"model":       relatedModel,
			"local_key":   localKey,
			"foreign_key": foreignKey,
		})

		return nil
	}, engine.SlotMeta{Description: "Define a one-to-many relationship."})

	// ORM.HASONE: 'Profile' { as: 'profile', foreign_key: 'user_id' }
	eng.Register("orm.hasOne", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		relatedModel := coerce.ToString(resolveValue(node.Value, scope))
		asName := strings.ToLower(relatedModel)
		localKey := "id"
		foreignKey := strings.ToLower(coerce.ToString(scope.GetDefault("_active_model", ""))) + "_id"

		for _, c := range node.Children {
			if c.Name == "as" {
				asName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "foreign_key" || c.Name == "fk" {
				foreignKey = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "local_key" || c.Name == "lk" {
				localKey = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		modelName := coerce.ToString(scope.GetDefault("_active_model", ""))
		if modelName == "" {
			return fmt.Errorf("orm.hasOne: no active model")
		}

		relKey := fmt.Sprintf("_rel_%s_%s", modelName, asName)
		scope.Set(relKey, map[string]interface{}{
			"type":        "hasOne",
			"model":       relatedModel,
			"local_key":   localKey,
			"foreign_key": foreignKey,
		})

		return nil
	}, engine.SlotMeta{Description: "Define a one-to-one relationship."})

	// ORM.BELONGSTOMANY: 'Role' { as: 'roles', table: 'role_user', foreign_pivot_key: 'user_id', related_pivot_key: 'role_id' }
	eng.Register("orm.belongsToMany", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		relatedModel := coerce.ToString(resolveValue(node.Value, scope))
		asName := strings.ToLower(relatedModel) + "s"

		modelName := coerce.ToString(scope.GetDefault("_active_model", ""))
		if modelName == "" {
			return fmt.Errorf("orm.belongsToMany: no active model")
		}

		// Defaults based on convention
		pivotTable := strings.ToLower(modelName) + "_" + strings.ToLower(relatedModel)
		foreignPivotKey := strings.ToLower(modelName) + "_id"
		relatedPivotKey := strings.ToLower(relatedModel) + "_id"
		parentKey := "id"
		relatedKey := "id"

		for _, c := range node.Children {
			if c.Name == "as" {
				asName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "table" {
				pivotTable = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "foreign_pivot_key" || c.Name == "fpk" {
				foreignPivotKey = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "related_pivot_key" || c.Name == "rpk" {
				relatedPivotKey = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "parent_key" || c.Name == "pk" {
				parentKey = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "related_key" || c.Name == "rk" {
				relatedKey = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		relKey := fmt.Sprintf("_rel_%s_%s", modelName, asName)
		scope.Set(relKey, map[string]interface{}{
			"type":              "belongsToMany",
			"model":             relatedModel,
			"table":             pivotTable,
			"foreign_pivot_key": foreignPivotKey,
			"related_pivot_key": relatedPivotKey,
			"parent_key":        parentKey,
			"related_key":       relatedKey,
		})

		return nil
	}, engine.SlotMeta{Description: "Define a many-to-many relationship."})

	// ORM.WITH: 'author' { orm.all: $posts }
	eng.Register("orm.with", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		relName := coerce.ToString(resolveValue(node.Value, scope))
		modelName := coerce.ToString(scope.GetDefault("_active_model", ""))

		relKey := fmt.Sprintf("_rel_%s_%s", modelName, relName)
		relRaw, ok := scope.Get(relKey)
		if !ok {
			return fmt.Errorf("orm.with: relationship '%s' not defined for model '%s'", relName, modelName)
		}
		rel := relRaw.(map[string]interface{})

		// 1. Execute the main query (usually the child block)
		if len(node.Children) == 0 {
			return fmt.Errorf("orm.with: expected child query block")
		}

		// We need to capture the result of the child query.
		// For simplicity, we assume the child query (like orm.all) sets a variable.
		// However, orm.all/db.get typically sets a variable in scope if 'as' is provided.

		err := eng.Execute(ctx, node.Children[0], scope)
		if err != nil {
			return err
		}

		// 2. Hydrate Relational Data
		// This is a simplified Eager Loading
		var targetVar string

		// If the child is directly 'set: $var_name', capture it. Otherwise look for 'as: $var_name' inside children
		if node.Children[0].Name == "set" {
			targetVar = strings.TrimPrefix(coerce.ToString(node.Children[0].Value), "$")
		} else {
			for _, c := range node.Children[0].Children {
				if c.Name == "as" {
					targetVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
					break
				}
			}
		}

		if targetVar == "" {
			return nil // No result to hydrate
		}

		data, ok := scope.Get(targetVar)
		if !ok || data == nil {
			return nil
		}

		// Hydration Logic (Eager Loading)
		if rel["type"] == "belongsTo" {
			foreignKey := rel["foreign_key"].(string)

			// 1. Extract all unique foreign keys from parents
			var fkVals []interface{}
			var parentObj map[string]interface{}
			var isList bool

			list, err := coerce.ToSlice(data)
			if err != nil {
				isList = false
				parentObj = data.(map[string]interface{})
				if fk, ok := parentObj[foreignKey]; ok && fk != nil {
					fkVals = append(fkVals, fk)
				}
			} else {
				isList = true
				fkMap := make(map[interface{}]bool)
				for _, item := range list {
					obj := item.(map[string]interface{})
					if fk, ok := obj[foreignKey]; ok && fk != nil {
						if !fkMap[fk] {
							fkMap[fk] = true
							fkVals = append(fkVals, fk)
						}
					}
				}
			}

			if len(fkVals) == 0 {
				return nil // Nothing to load
			}

			// 2. Fetch all related records in ONE query
			oldModel := scope.GetDefault("_active_model", "")
			oldQS, _ := scope.Get("_query_state")

			dialect := dbMgr.GetDialect("default")
			scope.Set("_active_model", rel["model"])
			scope.Set("_query_state", &QueryState{
				Table:   rel["model"].(string),
				Dialect: dialect,
				DBName:  "default",
			})

			eng.Execute(ctx, &engine.Node{
				Name: "db.where",
				Children: []*engine.Node{
					{Name: "col", Value: "id"}, // belongsTo usually targets `id` on the related table
					{Name: "op", Value: "IN"},
					{Name: "val", Value: fkVals},
				},
			}, scope)

			eng.Execute(ctx, &engine.Node{
				Name:     "db.get",
				Children: []*engine.Node{{Name: "as", Value: "___temp_rel_list"}},
			}, scope)

			// 3. Map related records back to parents in memory
			var relatedList []interface{}
			if temp, ok := scope.Get("___temp_rel_list"); ok && temp != nil {
				// We expect db.get to return a slice
				if slice, err := coerce.ToSlice(temp); err == nil {
					relatedList = slice
				}
			}

			// Build dictionary for O(1) matching
			relDict := make(map[interface{}]map[string]interface{})
			for _, relItem := range relatedList {
				r := relItem.(map[string]interface{})
				relDict[r["id"]] = r
			}

			if !isList {
				if fk, ok := parentObj[foreignKey]; ok {
					if matched, found := relDict[fk]; found {
						parentObj[relName] = matched
					}
				}
			} else {
				for _, item := range list {
					obj := item.(map[string]interface{})
					if fk, ok := obj[foreignKey]; ok {
						if matched, found := relDict[fk]; found {
							obj[relName] = matched
						}
					}
				}
			}

			// Restore context
			scope.Set("_active_model", oldModel)
			scope.Set("_query_state", oldQS)

		} else if rel["type"] == "hasMany" {
			localKey := rel["local_key"].(string)
			foreignKey := rel["foreign_key"].(string)

			// 1. Extract all unique local keys from parents
			var localKVals []interface{}
			var parentObj map[string]interface{}
			var isList bool

			list, err := coerce.ToSlice(data)
			if err != nil {
				isList = false
				parentObj = data.(map[string]interface{})
				if lk, ok := parentObj[localKey]; ok && lk != nil {
					localKVals = append(localKVals, lk)
				}
			} else {
				isList = true
				lkMap := make(map[interface{}]bool)
				for _, item := range list {
					obj := item.(map[string]interface{})
					if lk, ok := obj[localKey]; ok && lk != nil {
						if !lkMap[lk] {
							lkMap[lk] = true
							localKVals = append(localKVals, lk)
						}
					}
				}
			}

			if len(localKVals) == 0 {
				return nil
			}

			// 2. Fetch all related records in ONE query
			oldModel := scope.GetDefault("_active_model", "")
			oldQS, _ := scope.Get("_query_state")

			dialect := dbMgr.GetDialect("default")
			scope.Set("_active_model", rel["model"])
			scope.Set("_query_state", &QueryState{
				Table:   rel["model"].(string),
				Dialect: dialect,
				DBName:  "default",
			})

			eng.Execute(ctx, &engine.Node{
				Name: "db.where",
				Children: []*engine.Node{
					{Name: "col", Value: foreignKey},
					{Name: "op", Value: "IN"},
					{Name: "val", Value: localKVals},
				},
			}, scope)

			eng.Execute(ctx, &engine.Node{
				Name:     "db.get",
				Children: []*engine.Node{{Name: "as", Value: "___temp_rel_list"}},
			}, scope)

			// 3. Map related records back to parents in memory (Grouping by foreign key)
			var relatedList []interface{}
			if temp, ok := scope.Get("___temp_rel_list"); ok && temp != nil {
				if slice, err := coerce.ToSlice(temp); err == nil {
					relatedList = slice
				}
			}

			// Group related items by foreign_key
			relDict := make(map[interface{}][]map[string]interface{})
			for _, relItem := range relatedList {
				if r, ok := relItem.(map[string]interface{}); ok {
					fkVal := r[foreignKey]
					if fkVal != nil {
						// Ensure we don't end up with int64 / int mismatch in maps (sqlite returns int64)
						fkStr := coerce.ToString(fkVal)
						relDict[fkStr] = append(relDict[fkStr], r)
					}
				}
			}

			if !isList {
				if lk, ok := parentObj[localKey]; ok {
					lkStr := coerce.ToString(lk)
					if matched, found := relDict[lkStr]; found {
						parentObj[relName] = matched
					} else {
						parentObj[relName] = make([]map[string]interface{}, 0)
					}
				}
			} else {
				for _, item := range list {
					if obj, ok := item.(map[string]interface{}); ok {
						if lk, ok := obj[localKey]; ok {
							lkStr := coerce.ToString(lk)
							if matched, found := relDict[lkStr]; found {
								obj[relName] = matched
							} else {
								obj[relName] = make([]map[string]interface{}, 0)
							}
						}
					}
				}
			}

			// Restore context
			scope.Set("_active_model", oldModel)
			scope.Set("_query_state", oldQS)

		} else if rel["type"] == "hasOne" {
			localKey := rel["local_key"].(string)
			foreignKey := rel["foreign_key"].(string)

			// 1. Extract all unique local keys from parents
			var localKVals []interface{}
			var parentObj map[string]interface{}
			var isList bool

			list, err := coerce.ToSlice(data)
			if err != nil {
				isList = false
				parentObj = data.(map[string]interface{})
				if lk, ok := parentObj[localKey]; ok && lk != nil {
					localKVals = append(localKVals, lk)
				}
			} else {
				isList = true
				lkMap := make(map[interface{}]bool)
				for _, item := range list {
					obj := item.(map[string]interface{})
					if lk, ok := obj[localKey]; ok && lk != nil {
						if !lkMap[lk] {
							lkMap[lk] = true
							localKVals = append(localKVals, lk)
						}
					}
				}
			}

			if len(localKVals) == 0 {
				return nil
			}

			// 2. Fetch all related records in ONE query
			oldModel := scope.GetDefault("_active_model", "")
			oldQS, _ := scope.Get("_query_state")

			dialect := dbMgr.GetDialect("default")
			scope.Set("_active_model", rel["model"])
			scope.Set("_query_state", &QueryState{
				Table:   rel["model"].(string),
				Dialect: dialect,
				DBName:  "default",
			})

			eng.Execute(ctx, &engine.Node{
				Name: "db.where",
				Children: []*engine.Node{
					{Name: "col", Value: foreignKey},
					{Name: "op", Value: "IN"},
					{Name: "val", Value: localKVals},
				},
			}, scope)

			eng.Execute(ctx, &engine.Node{
				Name:     "db.get",
				Children: []*engine.Node{{Name: "as", Value: "___temp_rel_list"}},
			}, scope)

			// 3. Map related records back to parents in memory (1-to-1 matching)
			var relatedList []interface{}
			if temp, ok := scope.Get("___temp_rel_list"); ok && temp != nil {
				if slice, err := coerce.ToSlice(temp); err == nil {
					relatedList = slice
				}
			}

			// Map single item by foreign_key
			relDict := make(map[interface{}]map[string]interface{})
			for _, relItem := range relatedList {
				if r, ok := relItem.(map[string]interface{}); ok {
					fkVal := r[foreignKey]
					if fkVal != nil {
						fkStr := coerce.ToString(fkVal)
						// Only take the first one found for hasOne
						if _, exists := relDict[fkStr]; !exists {
							relDict[fkStr] = r
						}
					}
				}
			}

			if !isList {
				if lk, ok := parentObj[localKey]; ok {
					lkStr := coerce.ToString(lk)
					if matched, found := relDict[lkStr]; found {
						parentObj[relName] = matched
					} else {
						parentObj[relName] = nil
					}
				}
			} else {
				for _, item := range list {
					if obj, ok := item.(map[string]interface{}); ok {
						if lk, ok := obj[localKey]; ok {
							lkStr := coerce.ToString(lk)
							if matched, found := relDict[lkStr]; found {
								obj[relName] = matched
							} else {
								obj[relName] = nil
							}
						}
					}
				}
			}

			// Restore context
			scope.Set("_active_model", oldModel)
			scope.Set("_query_state", oldQS)
		} else if rel["type"] == "belongsToMany" {
			parentKey := rel["parent_key"].(string)
			foreignPivotKey := rel["foreign_pivot_key"].(string)

			// 1. Extract all unique parent keys from parents
			var parentKVals []interface{}
			var parentObj map[string]interface{}
			var isList bool

			list, err := coerce.ToSlice(data)
			if err != nil {
				isList = false
				parentObj = data.(map[string]interface{})
				if pk, ok := parentObj[parentKey]; ok && pk != nil {
					parentKVals = append(parentKVals, pk)
				}
			} else {
				isList = true
				pkMap := make(map[interface{}]bool)
				for _, item := range list {
					obj := item.(map[string]interface{})
					if pk, ok := obj[parentKey]; ok && pk != nil {
						if !pkMap[pk] {
							pkMap[pk] = true
							parentKVals = append(parentKVals, pk)
						}
					}
				}
			}

			if len(parentKVals) == 0 {
				return nil
			}

			// 2. Fetch all related records in ONE query
			oldModel := scope.GetDefault("_active_model", "")
			oldQS, _ := scope.Get("_query_state")

			dialect := dbMgr.GetDialect("default")
			scope.Set("_active_model", rel["model"])
			scope.Set("_query_state", &QueryState{
				Table:   rel["model"].(string),
				Dialect: dialect,
				DBName:  "default",
			})

			// db.join
			joinOn := []interface{}{
				fmt.Sprintf("%s.%s", rel["model"], rel["related_key"]),
				"=",
				fmt.Sprintf("%s.%s", rel["table"], rel["related_pivot_key"]),
			}
			eng.Execute(ctx, &engine.Node{
				Name: "db.join",
				Children: []*engine.Node{
					{Name: "table", Value: rel["table"]},
					{Name: "on", Value: joinOn},
				},
			}, scope)

			// db.where on pivot table
			whereCol := fmt.Sprintf("%s.%s", rel["table"], foreignPivotKey)
			eng.Execute(ctx, &engine.Node{
				Name: "db.where",
				Children: []*engine.Node{
					{Name: "col", Value: whereCol},
					{Name: "op", Value: "IN"},
					{Name: "val", Value: parentKVals},
				},
			}, scope)

			eng.Execute(ctx, &engine.Node{
				Name:     "db.get",
				Children: []*engine.Node{{Name: "as", Value: "___temp_rel_list"}},
			}, scope)

			// 3. Map related records back to parents in memory (Grouping by foreign_pivot_key)
			var relatedList []interface{}
			if temp, ok := scope.Get("___temp_rel_list"); ok && temp != nil {
				if slice, err := coerce.ToSlice(temp); err == nil {
					relatedList = slice
				}
			}

			// Group related items by foreign_pivot_key
			relDict := make(map[interface{}][]map[string]interface{})
			for _, relItem := range relatedList {
				if r, ok := relItem.(map[string]interface{}); ok {
					// In some SQL drivers, joined columns with identical names might overwrite one another
					// However, the foreign_pivot_key is usually unique to the pivot table and preserved.
					fkVal := r[foreignPivotKey]
					if fkVal != nil {
						fkStr := coerce.ToString(fkVal)
						relDict[fkStr] = append(relDict[fkStr], r)
					}
				}
			}

			if !isList {
				if pk, ok := parentObj[parentKey]; ok {
					pkStr := coerce.ToString(pk)
					if matched, found := relDict[pkStr]; found {
						parentObj[relName] = matched
					} else {
						parentObj[relName] = make([]map[string]interface{}, 0)
					}
				}
			} else {
				for _, item := range list {
					if obj, ok := item.(map[string]interface{}); ok {
						if pk, ok := obj[parentKey]; ok {
							pkStr := coerce.ToString(pk)
							if matched, found := relDict[pkStr]; found {
								obj[relName] = matched
							} else {
								obj[relName] = make([]map[string]interface{}, 0)
							}
						}
					}
				}
			}

			// Restore context
			scope.Set("_active_model", oldModel)
			scope.Set("_query_state", oldQS)
		}

		return nil
	}, engine.SlotMeta{Description: "Eager load a relationship."})

	// DB.SEED: { name: 'UserSeeder', data: [...] }
	eng.Register("db.seed", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Basic seeding: just execute the block
		for _, c := range node.Children {
			if err := eng.Execute(ctx, c, scope); err != nil {
				return err
			}
		}
		logNode := &engine.Node{Name: "log", Value: "🌱 Seeding completed."}
		eng.Execute(ctx, logNode, scope)
		return nil
	}, engine.SlotMeta{Description: "Execute database seeders."})
}
