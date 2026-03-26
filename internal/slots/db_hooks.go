package slots

import (
	"context"
	"sync"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// ==========================================
// DB HOOK REGISTRY
// ==========================================

type HookEvent string

const (
	HookBeforeInsert HookEvent = "before_insert"
	HookAfterInsert  HookEvent = "after_insert"
	HookBeforeUpdate HookEvent = "before_update"
	HookAfterUpdate  HookEvent = "after_update"
	HookBeforeDelete HookEvent = "before_delete"
	HookAfterDelete  HookEvent = "after_delete"
	HookBeforeSave   HookEvent = "before_save" // insert + update
	HookAfterSave    HookEvent = "after_save"  // insert + update
)

// hookEntry holds the node to execute for a specific event on a table
type hookEntry struct {
	node  *engine.Node
	scope *engine.Scope
}

var (
	hookMu       sync.RWMutex
	hookRegistry = make(map[string]map[HookEvent]*hookEntry) // table -> event -> node
)

// registerHook stores a hook node for a table/event pair
func registerHook(table string, event HookEvent, node *engine.Node, scope *engine.Scope) {
	hookMu.Lock()
	defer hookMu.Unlock()
	if _, ok := hookRegistry[table]; !ok {
		hookRegistry[table] = make(map[HookEvent]*hookEntry)
	}
	hookRegistry[table][event] = &hookEntry{node: node, scope: scope}
}

// fireHook executes a hook if one is registered for the given table and event.
// It injects $data and $table into the hook's scope before execution.
func fireHook(ctx context.Context, eng *engine.Engine, table string, event HookEvent, data interface{}, scope *engine.Scope) error {
	hookMu.RLock()
	events, ok := hookRegistry[table]
	if !ok {
		hookMu.RUnlock()
		return nil
	}
	entry, ok := events[event]
	hookMu.RUnlock()
	if !ok {
		return nil
	}

	// Create a child scope so the hook can read request variables but
	// any new variables it sets don't pollute the caller's scope.
	hookScope := engine.NewScope(scope)
	hookScope.Set("data", data)
	hookScope.Set("table", table)

	for _, child := range entry.node.Children {
		if child.Name == string(event) {
			return eng.Execute(ctx, child, hookScope)
		}
	}
	return nil
}

// RegisterDBHookSlots registers the db.hook slot
func RegisterDBHookSlots(eng *engine.Engine) {
	// ==========================================
	// SLOT: DB.HOOK
	// ==========================================
	eng.Register("db.hook", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		table := coerce.ToString(resolveValue(node.Value, scope))
		if table == "" {
			// Try children for table name
			for _, c := range node.Children {
				if c.Name == "table" {
					table = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}
		if table == "" {
			return nil
		}

		// Register hooks for supported events
		events := []HookEvent{
			HookBeforeInsert, HookAfterInsert,
			HookBeforeUpdate, HookAfterUpdate,
			HookBeforeDelete, HookAfterDelete,
			HookBeforeSave, HookAfterSave,
		}
		for _, event := range events {
			for _, c := range node.Children {
				if c.Name == string(event) {
					registerHook(table, event, node, scope)
					break
				}
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Register lifecycle hooks for a database table (before/after insert, update, delete, save).",
		Example: `db.hook: 'posts' {
  before_insert: {
    var: $data.slug slug($data.title)
  }
  after_save: {
    cache.forget: "posts_list"
  }
  after_update: {
    db.insert: activity_log { action: "updated" table: "posts" }
  }
}`,
		Inputs: map[string]engine.InputMeta{
			"before_insert": {Description: "Code block executed before an INSERT", Required: false},
			"after_insert":  {Description: "Code block executed after an INSERT", Required: false},
			"before_update": {Description: "Code block executed before an UPDATE", Required: false},
			"after_update":  {Description: "Code block executed after an UPDATE", Required: false},
			"before_delete": {Description: "Code block executed before a DELETE", Required: false},
			"after_delete":  {Description: "Code block executed after a DELETE", Required: false},
			"before_save":   {Description: "Code block executed before INSERT or UPDATE", Required: false},
			"after_save":    {Description: "Code block executed after INSERT or UPDATE", Required: false},
		},
	})
}
