package slots

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"testing"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/pbkdf2"
)

func TestAspNetSlots(t *testing.T) {
	// Setup DB
	dbMgr := dbmanager.NewDBManager()
	err := dbMgr.AddConnection("default", "sqlite", ":memory:", 1, 1)
	if err != nil {
		t.Fatalf("Failed to create in-memory db: %v", err)
	}
	defer dbMgr.Close()

	// Seed DB
	db := dbMgr.GetConnection("default")
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
	RegisterAspNetSlots(eng, dbMgr)

	t.Run("aspnet.login success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "aspnet.login",
			Children: []*engine.Node{
				{Name: "username", Value: "asp@example.com"},
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

	t.Run("aspnet.hash and aspnet.verify slots", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name:  "aspnet.hash",
			Value: "AspPass123!",
			Children: []*engine.Node{
				{Name: "iterations", Value: int64(10000)},
				{Name: "as", Value: "$hashed"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		hashedRaw, ok := scope.Get("hashed")
		assert.True(t, ok)
		hashedStr := hashedRaw.(string)
		assert.NotEmpty(t, hashedStr)

		// Verify using verify slot
		verifyNode := &engine.Node{
			Name: "aspnet.verify",
			Children: []*engine.Node{
				{Name: "hash", Value: hashedStr},
				{Name: "password", Value: "AspPass123!"},
				{Name: "as", Value: "$is_valid"},
			},
		}

		err = eng.Execute(context.Background(), verifyNode, scope)
		assert.NoError(t, err)

		isValid, ok := scope.Get("is_valid")
		assert.True(t, ok)
		assert.True(t, isValid.(bool))
	})

	t.Run("aspnet.validate_password custom policy", func(t *testing.T) {
		scope := engine.NewScope(nil)

		// 1. Weak password validation
		node := &engine.Node{
			Name: "aspnet.validate_password",
			Children: []*engine.Node{
				{Name: "password", Value: "123"},
				{Name: "required_length", Value: int64(6)},
				{Name: "as", Value: "$is_ok"},
				{Name: "errors_as", Value: "$errs"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		isOk, ok := scope.Get("is_ok")
		assert.True(t, ok)
		assert.False(t, isOk.(bool))

		errsRaw, ok := scope.Get("errs")
		assert.True(t, ok)
		errs := errsRaw.([]interface{})
		assert.NotEmpty(t, errs)

		// 2. Strong password validation
		scopeStrong := engine.NewScope(nil)
		nodeStrong := &engine.Node{
			Name: "aspnet.validate_password",
			Children: []*engine.Node{
				{Name: "password", Value: "P@ssw0rd123!"},
				{Name: "required_length", Value: int64(8)},
				{Name: "as", Value: "$is_ok"},
				{Name: "errors_as", Value: "$errs"},
			},
		}

		err = eng.Execute(context.Background(), nodeStrong, scopeStrong)
		assert.NoError(t, err)

		isOkStrong, ok := scopeStrong.Get("is_ok")
		assert.True(t, ok)
		assert.True(t, isOkStrong.(bool))

		errsStrong, ok := scopeStrong.Get("errs")
		assert.True(t, ok)
		assert.Nil(t, errsStrong)
	})
}

func TestVerifyAspNetHash(t *testing.T) {
	garbage := "not_a_valid_base64"
	if VerifyAspNetHash(garbage, "pass") {
		t.Error("Expected false for garbage input")
	}

	validBase64ButShort := base64.StdEncoding.EncodeToString([]byte("short"))
	if VerifyAspNetHash(validBase64ButShort, "pass") {
		t.Error("Expected false for short input")
	}

	blob := make([]byte, 58)
	blob[0] = 0x01
	blob[12] = 0x10

	hash := base64.StdEncoding.EncodeToString(blob)
	if VerifyAspNetHash(hash, "wrong") {
		t.Error("Expected false for mismatch password")
	}
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
