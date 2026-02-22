package slots

import (
	"context"
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

		if isUpdate {
			// Update Logic
			originalWhere := qs.Where
			qs.Where = append(qs.Where, WhereCond{Column: primaryKey, Op: "=", Value: idVal})
			
			// Build sets excluding the PK
			updateNode := &engine.Node{Name: "db.update"}
			for k, v := range data {
				if k == primaryKey { continue }
				updateNode.Children = append(updateNode.Children, &engine.Node{Name: k, Value: v})
			}
			
			err := eng.Execute(ctx, updateNode, scope)
			qs.Where = originalWhere
			return err
		} else {
			// Insert Logic
			insertNode := &engine.Node{Name: "db.insert"}
			for k, v := range data {
				insertNode.Children = append(insertNode.Children, &engine.Node{Name: k, Value: v})
			}
			err := eng.Execute(ctx, insertNode, scope)
			if err == nil {
				// Set the new ID back to the object if possible
				if lastId, ok := scope.Get("db_last_id"); ok {
					data[primaryKey] = lastId
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
		// This is a simplified Eager Loading (not full N+1 fix yet, but functional API)
		// We'll look for the result in scope based on the child slot's 'as' attribute.
		var targetVar string
		for _, c := range node.Children[0].Children {
			if c.Name == "as" {
				targetVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if targetVar == "" {
			return nil // No result to hydrate
		}

		data, ok := scope.Get(targetVar)
		if !ok {
			return nil
		}

		// Hydration Logic
		if rel["type"] == "belongsTo" {
			// Find related for each item
			list, err := coerce.ToSlice(data)
			if err != nil {
				// Single object
				obj := data.(map[string]interface{})
				fkVal := obj[rel["foreign_key"].(string)]
				// Temporary switch model context
				oldModel := scope.GetDefault("_active_model", "")
				oldQS, _ := scope.Get("_query_state")
				
				// Set new model context
				dialect := dbMgr.GetDialect("default") // Assuming default DB
				scope.Set("_active_model", rel["model"])
				scope.Set("_query_state", &QueryState{
					Table: rel["model"].(string),
					Dialect: dialect,
					DBName: "default",
				})

				eng.Execute(ctx, &engine.Node{
					Name: "orm.find",
					Value: fkVal,
					Children: []*engine.Node{{Name: "as", Value: relName}},
				}, scope)
				
				if related, ok := scope.Get(relName); ok {
					obj[relName] = related
				}
				
				// Restore context
				scope.Set("_active_model", oldModel)
				scope.Set("_query_state", oldQS)
			} else {
				// List of objects
				for _, item := range list {
					obj := item.(map[string]interface{})
					fkVal := obj[rel["foreign_key"].(string)]
					
					oldModel := scope.GetDefault("_active_model", "")
					oldQS, _ := scope.Get("_query_state")
					
					dialect := dbMgr.GetDialect("default")
					scope.Set("_active_model", rel["model"])
					scope.Set("_query_state", &QueryState{
						Table: rel["model"].(string),
						Dialect: dialect,
						DBName: "default",
					})

					eng.Execute(ctx, &engine.Node{
						Name: "orm.find",
						Value: fkVal,
						Children: []*engine.Node{{Name: "as", Value: "___temp_rel"}},
					}, scope)
					if related, ok := scope.Get("___temp_rel"); ok {
						obj[relName] = related
					}
					
					scope.Set("_active_model", oldModel)
					scope.Set("_query_state", oldQS)
				}
			}
		} else if rel["type"] == "hasMany" {
			// Find many for each item
			list, err := coerce.ToSlice(data)
			if err != nil {
				// Single object
				obj := data.(map[string]interface{})
				localVal := obj[rel["local_key"].(string)]
				
				oldModel := scope.GetDefault("_active_model", "")
				oldQS, _ := scope.Get("_query_state")

				dialect := dbMgr.GetDialect("default")
				scope.Set("_active_model", rel["model"])
				scope.Set("_query_state", &QueryState{
					Table: rel["model"].(string),
					Dialect: dialect,
					DBName: "default",
				})
				
				eng.Execute(ctx, &engine.Node{
					Name: "db.where",
					Children: []*engine.Node{
						{Name: "col", Value: rel["foreign_key"]},
						{Name: "op", Value: "="},
						{Name: "val", Value: localVal},
					},
				}, scope)
				eng.Execute(ctx, &engine.Node{
					Name: "db.get",
					Children: []*engine.Node{{Name: "as", Value: "___temp_rel_list"}},
				}, scope)
				
				if related, ok := scope.Get("___temp_rel_list"); ok {
					obj[relName] = related
				}
				
				scope.Set("_active_model", oldModel)
				scope.Set("_query_state", oldQS)
			} else {
				for _, item := range list {
					obj := item.(map[string]interface{})
					localVal := obj[rel["local_key"].(string)]
					
					oldModel := scope.GetDefault("_active_model", "")
					oldQS, _ := scope.Get("_query_state")

					dialect := dbMgr.GetDialect("default")
					scope.Set("_active_model", rel["model"])
					scope.Set("_query_state", &QueryState{
						Table: rel["model"].(string),
						Dialect: dialect,
						DBName: "default",
					})
					
					eng.Execute(ctx, &engine.Node{
						Name: "db.where",
						Children: []*engine.Node{
							{Name: "col", Value: rel["foreign_key"]},
							{Name: "op", Value: "="},
							{Name: "val", Value: localVal},
						},
					}, scope)
					eng.Execute(ctx, &engine.Node{
						Name: "db.get",
						Children: []*engine.Node{{Name: "as", Value: "___temp_rel_list"}},
					}, scope)
					
					if related, ok := scope.Get("___temp_rel_list"); ok {
						obj[relName] = related
					}
					
					scope.Set("_active_model", oldModel)
					scope.Set("_query_state", oldQS)
				}
			}
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
