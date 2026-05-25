package slots

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

func TestAuthSlots(t *testing.T) {
	// Setup DB
	dbMgr := dbmanager.NewDBManager()
	// Use in-memory SQLite
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	if err != nil {
		t.Fatalf("Failed to create in-memory db: %v", err)
	}
	defer dbMgr.Close()

	// Seed DB
	db := dbMgr.GetConnection("default")
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT, username TEXT, password TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	_, err = db.Exec(`INSERT INTO users (email, username, password) VALUES (?, ?, ?)`, "test@example.com", "testuser", string(hashed))
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE AspNetUsers (Id TEXT PRIMARY KEY, UserName TEXT, NormalizedUserName TEXT, Email TEXT, NormalizedEmail TEXT, PasswordHash TEXT, TenantId TEXT, FullName TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create AspNetUsers table: %v", err)
	}

	aspnetHash := generateAspNetHash("AspPass123!")
	_, err = db.Exec(`INSERT INTO AspNetUsers (Id, UserName, NormalizedUserName, Email, NormalizedEmail, PasswordHash, TenantId, FullName) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"asp-user-uuid-999", "AspUser", "ASPUSER", "asp@example.com", "ASP@EXAMPLE.COM", aspnetHash, "tenant-abc", "ASP.NET Migrated User")
	if err != nil {
		t.Fatalf("Failed to insert aspnet user: %v", err)
	}

	eng := engine.NewEngine()
	RegisterAuthSlots(eng, dbMgr)

	t.Run("auth.login success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "auth.login",
			Children: []*engine.Node{
				{Name: "username", Value: "testuser"}, // Passing username
				{Name: "col_user", Value: "username"}, // Tell it to check username column
				{Name: "password", Value: "secret"},   // Correct password
				{Name: "as", Value: "$token"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		token, ok := scope.Get("token")
		assert.True(t, ok)
		assert.NotEmpty(t, token)
	})

	t.Run("auth.login fail", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "auth.login",
			Children: []*engine.Node{
				{Name: "username", Value: "testuser"},
				{Name: "password", Value: "wrongpass"},
				{Name: "as", Value: "$token"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("jwt.sign and verify", func(t *testing.T) {
		scope := engine.NewScope(nil)

		// Sign
		nodeSign := &engine.Node{
			Name: "jwt.sign",
			Children: []*engine.Node{
				{Name: "secret", Value: "mysecret"},
				{Name: "claims", Value: map[string]interface{}{"sub": "123"}},
				{Name: "as", Value: "$token"},
			},
		}
		err := eng.Execute(context.Background(), nodeSign, scope)
		assert.NoError(t, err)

		// Verify
		nodeVerify := &engine.Node{
			Name: "jwt.verify",
			Children: []*engine.Node{
				{Name: "token", Value: "$token"},
				{Name: "secret", Value: "mysecret"},
				{Name: "as", Value: "$claims"},
			},
		}
		err = eng.Execute(context.Background(), nodeVerify, scope)
		assert.NoError(t, err)

		claimsRaw, ok := scope.Get("claims")
		assert.True(t, ok)
		claims, ok := claimsRaw.(jwt.MapClaims)
		assert.True(t, ok, "Expected jwt.MapClaims")
		assert.Equal(t, "123", claims["sub"])
	})

	// Test auth.user with manually mocked scope
	t.Run("auth.user from session", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("session", map[string]interface{}{"user_id": 1})

		node := &engine.Node{
			Name: "auth.user",
			Children: []*engine.Node{
				{Name: "as", Value: "$u"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		u, _ := scope.Get("u")
		uMap := u.(map[string]interface{})
		assert.Equal(t, 1, uMap["user_id"])
	})

	t.Run("auth.aspnet_login success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "auth.aspnet_login",
			Children: []*engine.Node{
				{Name: "username", Value: "asp@example.com"}, // match via email
				{Name: "password", Value: "AspPass123!"},
				{Name: "fields", Value: []interface{}{"TenantId", "FullName"}},
				{Name: "as", Value: "$token"},
				{Name: "user_as", Value: "$usr"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		token, ok := scope.Get("token")
		assert.True(t, ok)
		assert.NotEmpty(t, token)

		usrRaw, ok := scope.Get("usr")
		assert.True(t, ok)
		usr, ok := usrRaw.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "asp-user-uuid-999", usr["id"])
		assert.Equal(t, "AspUser", usr["username"])
		assert.Equal(t, "asp@example.com", usr["email"])
		assert.Equal(t, "tenant-abc", usr["TenantId"])
		assert.Equal(t, "ASP.NET Migrated User", usr["FullName"])
	})
}

func generateAspNetHash(password string) string {
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = byte(i)
	}
	iterCount := 10000
	subkey := pbkdf2.Key([]byte(password), salt, iterCount, 32, sha256.New)

	blob := make([]byte, 1+4+4+4+16+32)
	blob[0] = 0x01                          // Version V3
	binary.BigEndian.PutUint32(blob[1:5], 1) // PRF = SHA256
	binary.BigEndian.PutUint32(blob[5:9], uint32(iterCount))
	binary.BigEndian.PutUint32(blob[9:13], 16) // saltLen = 16
	copy(blob[13:29], salt)
	copy(blob[29:], subkey)

	return base64.StdEncoding.EncodeToString(blob)
}
