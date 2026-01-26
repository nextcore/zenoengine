package slots

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

func RegisterSecuritySlots(eng *engine.Engine) {

	// 1. CRYPTO.HASH
	eng.Register("crypto.hash", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Value utama (shorthand): crypto.hash: $password
		input := coerce.ToString(resolveValue(node.Value, scope))
		target := "hash_result"

		for _, c := range node.Children {
			if c.Name == "text" || c.Name == "val" {
				input = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if input == "" {
			return fmt.Errorf("crypto.hash: input text is required")
		}

		bytes, err := bcrypt.GenerateFromPassword([]byte(input), 10)
		if err != nil {
			return err
		}

		scope.Set(target, string(bytes))
		return nil
	}, engine.SlotMeta{Example: "crypto.hash: $pass\n  as: $hashed"})

	// 2. CRYPTO.VERIFY
	eng.Register("crypto.verify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var hash, text string
		target := "verify_result"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "hash" {
				hash = coerce.ToString(val)
			}
			if c.Name == "text" {
				text = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(text))
		isValid := (err == nil)
		scope.Set(target, isValid)
		return nil
	}, engine.SlotMeta{Example: "crypto.verify\n  hash: $h\n  text: $p"})

	// 3. SEC.CSRF_TOKEN
	eng.Register("sec.csrf_token", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return fmt.Errorf("httpRequest not found in context")
		}
		r := reqVal.(*http.Request)
		token := csrf.Token(r)

		target := "csrf_token"
		if node.Value != nil {
			target = strings.TrimPrefix(coerce.ToString(node.Value), "$")
		}

		scope.Set(target, token)
		return nil
	}, engine.SlotMeta{Example: "sec.csrf_token: $token"})

	// 4. CRYPTO.VERIFY_ASPNET (Identity V3)
	eng.Register("crypto.verify_aspnet", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var hash, password string
		target := "verify_result"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "hash" {
				hash = coerce.ToString(val)
			}
			if c.Name == "password" {
				password = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if hash == "" || password == "" {
			return fmt.Errorf("crypto.verify_aspnet: hash and password are required")
		}

		isValid := VerifyAspNetHash(hash, password)
		scope.Set(target, isValid)
		return nil
	}, engine.SlotMeta{Example: "crypto.verify_aspnet\n  hash: $db_hash\n  password: $input_pass"})
}

// VerifyAspNetHash verifies ASP.NET Identity V3 password hashes
// Format: 0x01 (1 byte) | Prf (1 byte) | IterCount (4 bytes) | SaltLen (4 bytes) | Salt | Subkey
func VerifyAspNetHash(hashedPassword, providedPassword string) bool {
	decoded, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}

	if len(decoded) < 13 {
		return false // Too short to be valid V3
	}

	// Verify version byte (0x01)
	if decoded[0] != 0x01 {
		return false // Not Identity V3
	}

	// Read header info
	// prf := decoded[1] // 0=SHA1, 1=SHA256, 2=SHA512. We assume SHA256 (1) for Identity V3 default.
	iterCount := int(binary.BigEndian.Uint32(decoded[5:9])) // Note: ASP.NET uses BigEndian for these ints in the binary blob
	saltLen := int(binary.BigEndian.Uint32(decoded[9:13]))

	if len(decoded) < 13+saltLen {
		return false
	}

	salt := decoded[13 : 13+saltLen]
	expectedSubkey := decoded[13+saltLen:]

	// Hash the provided password with the same parameters
	// PRF: HMAC-SHA256 (default for V3)
	// KeyLen: 32 bytes (256 bits)
	dk := pbkdf2.Key([]byte(providedPassword), salt, iterCount, 32, sha256.New)

	return subtle.ConstantTimeCompare(dk, expectedSubkey) == 1
}
