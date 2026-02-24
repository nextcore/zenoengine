package slots

import (
	"context"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestORMSlots(t *testing.T) {
	// Setup DB
	dbMgr := dbmanager.NewDBManager()
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	if err != nil {
		t.Fatalf("Failed to create in-memory db: %v", err)
	}
	defer dbMgr.Close()

	// Seed DB
	db := dbMgr.GetConnection("default")
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	eng := engine.NewEngine()
	// Register dependencies
	RegisterDBSlots(eng, dbMgr)
	RegisterRawDBSlots(eng, dbMgr)
	RegisterORMSlots(eng, dbMgr)

	t.Run("orm.model sets query state", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name:  "orm.model",
			Value: "users",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		qsRaw, ok := scope.Get("_query_state")
		assert.True(t, ok)
		qs := qsRaw.(*QueryState)
		assert.Equal(t, "users", qs.Table)

		modelName, _ := scope.Get("_active_model")
		assert.Equal(t, "users", modelName)
	})

	t.Run("orm.find", func(t *testing.T) {
		scope := engine.NewScope(nil)
		// Set model first
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		node := &engine.Node{
			Name:  "orm.find",
			Value: 1,
			Children: []*engine.Node{
				{Name: "as", Value: "$u"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		uRaw, ok := scope.Get("u")
		assert.True(t, ok)
		u := uRaw.(map[string]interface{})
		assert.Equal(t, "Alice", u["name"])
	})

	t.Run("orm.save insert", func(t *testing.T) {
		scope := engine.NewScope(nil)
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		data := map[string]interface{}{
			"name":  "Bob",
			"email": "bob@example.com",
		}
		scope.Set("new_user", data)

		node := &engine.Node{
			Name:  "orm.save",
			Value: "$new_user",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Check if inserted
		// Manually check via DB or slot
		// Let's use db.count
		scope.Set("total", 0)
		eng.Execute(context.Background(), &engine.Node{Name: "db.count", Children: []*engine.Node{{Name: "as", Value: "$total"}}}, scope)
		total, _ := scope.Get("total")
		assert.Equal(t, 2, total)
	})

	t.Run("orm.save update", func(t *testing.T) {
		scope := engine.NewScope(nil)
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		data := map[string]interface{}{
			"id":   1,
			"name": "Alice Updated",
		}
		scope.Set("update_user", data)

		node := &engine.Node{
			Name:  "orm.save",
			Value: "$update_user",
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Verify update
		eng.Execute(context.Background(), &engine.Node{
			Name:     "orm.find",
			Value:    1,
			Children: []*engine.Node{{Name: "as", Value: "$u2"}}},
			scope)

		u2Raw, _ := scope.Get("u2")
		u2 := u2Raw.(map[string]interface{})
		assert.Equal(t, "Alice Updated", u2["name"])
	})

	t.Run("orm.save fillable protection", func(t *testing.T) {
		scope := engine.NewScope(nil)
		eng.Execute(context.Background(), &engine.Node{Name: "orm.model", Value: "users"}, scope)

		// Add column 'is_admin' for this test
		db := dbMgr.GetConnection("default")
		db.Exec(`ALTER TABLE users ADD COLUMN is_admin INTEGER DEFAULT 0`)

		scope.Set("malicious_user", map[string]interface{}{
			"name":     "Eve",
			"email":    "eve@example.com",
			"is_admin": 1,
		})

		script := `
			orm.model: 'users' {
				fillable: 'name, email'
			}

			orm.model: 'users'
			orm.save: $malicious_user
		`
		node, err := engine.ParseString(script, "orm_fillable_test.zl")
		assert.NoError(t, err)

		err = eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Assert using native DB connection
		var name, email string
		var isAdmin int
		err = db.QueryRow("SELECT name, email, is_admin FROM users WHERE name = 'Eve'").Scan(&name, &email, &isAdmin)
		assert.NoError(t, err)

		assert.Equal(t, "Eve", name)
		assert.Equal(t, "eve@example.com", email)
		assert.Equal(t, 0, isAdmin) // Should be 0 (default) because mass assignment blocked it
	})

	t.Run("orm.with eager loading", func(t *testing.T) {
		// Setup related table
		db := dbMgr.GetConnection("default")
		_, err := db.Exec(`CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT)`)
		assert.NoError(t, err)

		_, err = db.Exec(`INSERT INTO posts (user_id, title) VALUES (?, ?), (?, ?)`, 1, "Post A", 1, "Post B")
		assert.NoError(t, err)

		scope := engine.NewScope(nil)

		// Define users model with hasMany posts and execute eager loading
		script := `
			// 1. Define Model
			orm.model: 'users' {
				orm.hasMany: 'posts' {
					as: 'posts'
					foreign_key: 'user_id'
					local_key: 'id'
				}
			}
			
			// 2. Fetch Users
			orm.model: 'users'
			db.get { as: $userList }
			
			// 3. Hydrate Users
			orm.model: 'users'
			orm.with: 'posts' {
				set: $userList { val: $userList }
			}
		`

		node, err := engine.ParseString(script, "orm_with_test.zl")
		assert.NoError(t, err)

		err = eng.Execute(context.Background(), node, scope)
		if err != nil {
			t.Log(err)
		}

		usersRaw, ok := scope.Get("userList")
		assert.True(t, ok)

		users := usersRaw.([]map[string]interface{})
		assert.Greater(t, len(users), 0)

		alice := users[0]
		postsRaw, hasPosts := alice["posts"]
		assert.True(t, hasPosts)

		posts := postsRaw.([]map[string]interface{})
		assert.Equal(t, 2, len(posts))

		firstPost := posts[0]
		assert.Equal(t, "Post A", firstPost["title"])
	})

	t.Run("orm.with hasOne eager loading", func(t *testing.T) {
		// Setup related table
		db := dbMgr.GetConnection("default")
		_, err := db.Exec(`CREATE TABLE profiles (id INTEGER PRIMARY KEY, user_id INTEGER, bio TEXT)`)
		assert.NoError(t, err)

		_, err = db.Exec(`INSERT INTO profiles (user_id, bio) VALUES (?, ?)`, 1, "Alice Bio")
		assert.NoError(t, err)

		scope := engine.NewScope(nil)

		// Define users model with hasOne profile and execute eager loading
		script := `
			// 1. Define Model
			orm.model: 'users' {
				orm.hasOne: 'profiles' {
					as: 'profile'
					foreign_key: 'user_id'
					local_key: 'id'
				}
			}
			
			// 2. Fetch Users
			orm.model: 'users'
			db.get { as: $userList }
			
			// 3. Hydrate Users
			orm.model: 'users'
			orm.with: 'profile' {
				set: $userList { val: $userList }
			}
		`

		node, err := engine.ParseString(script, "orm_hasone_test.zl")
		assert.NoError(t, err)

		err = eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		usersRaw, ok := scope.Get("userList")
		assert.True(t, ok)

		users := usersRaw.([]map[string]interface{})
		assert.Greater(t, len(users), 0)

		alice := users[0]
		profileRaw, hasProfile := alice["profile"]
		assert.True(t, hasProfile)

		profile := profileRaw.(map[string]interface{})
		assert.Equal(t, "Alice Bio", profile["bio"])
	})

	t.Run("orm.with belongsToMany eager loading", func(t *testing.T) {
		db := dbMgr.GetConnection("default")
		_, err := db.Exec(`CREATE TABLE roles (id INTEGER PRIMARY KEY, name TEXT)`)
		assert.NoError(t, err)
		_, err = db.Exec(`CREATE TABLE role_user (user_id INTEGER, role_id INTEGER)`)
		assert.NoError(t, err)

		_, err = db.Exec(`INSERT INTO roles (name) VALUES (?), (?)`, "Admin", "Editor")
		assert.NoError(t, err)

		// User 1 gets Role 1 and Role 2
		_, err = db.Exec(`INSERT INTO role_user (user_id, role_id) VALUES (?, ?), (?, ?)`, 1, 1, 1, 2)
		assert.NoError(t, err)

		scope := engine.NewScope(nil)

		script := `
			// 1. Define Model
			orm.model: 'users' {
				orm.belongsToMany: 'roles' {
					as: 'roles'
					table: 'role_user'
					foreign_pivot_key: 'user_id'
					related_pivot_key: 'role_id'
				}
			}
			
			// 2. Fetch Users
			orm.model: 'users'
			db.get { as: $userList }
			
			// 3. Hydrate Users
			orm.model: 'users'
			orm.with: 'roles' {
				set: $userList { val: $userList }
			}
		`

		node, err := engine.ParseString(script, "orm_belongstomany_test.zl")
		assert.NoError(t, err)

		err = eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		usersRaw, ok := scope.Get("userList")
		assert.True(t, ok)

		users := usersRaw.([]map[string]interface{})
		assert.Greater(t, len(users), 0)

		alice := users[0]
		rolesRaw, hasRoles := alice["roles"]
		assert.True(t, hasRoles)

		roles := rolesRaw.([]map[string]interface{})
		assert.Equal(t, 2, len(roles))

		assert.Equal(t, "Admin", roles[0]["name"])
		assert.Equal(t, "Editor", roles[1]["name"])
	})

	t.Run("orm and db.query integration", func(t *testing.T) {
		db := dbMgr.GetConnection("default")
		_, err := db.Exec(`CREATE TABLE teams (id INTEGER PRIMARY KEY, name TEXT)`)
		assert.NoError(t, err)

		_, err = db.Exec(`INSERT INTO teams (name) VALUES (?)`, "Engineering")
		assert.NoError(t, err)

		// Add team_id to users
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN team_id INTEGER DEFAULT 1`)
		_, err = db.Exec(`UPDATE users SET team_id = 1`)

		scope := engine.NewScope(nil)

		script := `
			orm.model: 'users' {
				orm.belongsTo: 'teams' {
					as: 'team'
					foreign_key: 'team_id'
				}
			}

			orm.model: 'users'
			db.where {
				col: 'team_id'
				val: 1
			}
			db.get { as: $members }

			orm.model: 'users'
			orm.with: 'team' {
				set: $members { val: $members }
			}
		`

		node, err := engine.ParseString(script, "orm_integration_test.zl")
		assert.NoError(t, err)

		err = eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		membersRaw, ok := scope.Get("members")
		assert.True(t, ok)

		members := membersRaw.([]map[string]interface{})
		assert.Greater(t, len(members), 0)

		teamRaw, hasTeam := members[0]["team"]
		assert.True(t, hasTeam)
		team := teamRaw.(map[string]interface{})
		assert.Equal(t, "Engineering", team["name"])
	})
}
