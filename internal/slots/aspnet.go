package slots

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/pbkdf2"
)

// RegisterAspNetSlots registers slots for ASP.NET Identity migration
func RegisterAspNetSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// 1. ASPNET.LOGIN
	eng.Register("aspnet.login", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var username, password string
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "458127c2cffdd41a448b5d37b825188bf12db10e5c98cb03b681da667ac3b294_pekalongan_kota_2025_!@#_jgn_disebar" // Default fallback
		}
		targetToken := "token"
		targetUser := "user"
		dbName := "default"
		expiresIn := int64(86400) // 24 hours default
		var customFields []string

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "username" || c.Name == "email" {
				username = coerce.ToString(val)
			}
			if c.Name == "password" {
				password = coerce.ToString(val)
			}
			if c.Name == "secret" {
				jwtSecret = coerce.ToString(val)
			}
			if c.Name == "as" {
				targetToken = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "user_as" {
				targetUser = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "db" {
				dbName = coerce.ToString(val)
			}
			if c.Name == "expires_in" {
				expiresIn, _ = coerce.ToInt64(val)
			}
			if c.Name == "fields" {
				if fieldsVal := parseNodeValue(c, scope); fieldsVal != nil {
					if list, ok := fieldsVal.([]interface{}); ok {
						for _, item := range list {
							customFields = append(customFields, coerce.ToString(item))
						}
					}
				}
			}
		}

		if username == "" || password == "" {
			return fmt.Errorf("aspnet.login: username and password required")
		}

		// DB Lookup
		db := dbMgr.GetConnection(dbName)
		dialect := dbMgr.GetDialect(dbName)
		if db == nil {
			return fmt.Errorf("aspnet.login: database connection '%s' not found", dbName)
		}

		// Normalize search parameters
		upperUsername := strings.ToUpper(username)

		// Build columns select slice dynamically
		columns := []string{
			dialect.QuoteIdentifier("Id"),
			dialect.QuoteIdentifier("UserName"),
			dialect.QuoteIdentifier("Email"),
			dialect.QuoteIdentifier("PasswordHash"),
		}
		for _, cf := range customFields {
			columns = append(columns, dialect.QuoteIdentifier(cf))
		}

		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s OR %s = %s OR %s = %s OR %s = %s %s",
			strings.Join(columns, ", "),
			dialect.QuoteIdentifier("AspNetUsers"),
			dialect.QuoteIdentifier("NormalizedUserName"), dialect.Placeholder(1),
			dialect.QuoteIdentifier("NormalizedEmail"), dialect.Placeholder(2),
			dialect.QuoteIdentifier("UserName"), dialect.Placeholder(3),
			dialect.QuoteIdentifier("Email"), dialect.Placeholder(4),
			dialect.Limit(1, 0))

		var dbID interface{}
		var dbUsername, dbEmail, dbPass string

		scanTargets := make([]interface{}, 4+len(customFields))
		scanTargets[0] = &dbID
		scanTargets[1] = &dbUsername
		scanTargets[2] = &dbEmail
		scanTargets[3] = &dbPass

		customValues := make([]interface{}, len(customFields))
		for i := range customFields {
			scanTargets[4+i] = &customValues[i]
		}

		err := db.QueryRowContext(ctx, query, upperUsername, upperUsername, username, username).Scan(scanTargets...)
		if err != nil {
			return fmt.Errorf("aspnet.login: invalid credentials")
		}

		// Verify Password using ASP.NET Identity V3 PBKDF2/HMAC-SHA256
		if !VerifyAspNetHash(dbPass, password) {
			return fmt.Errorf("aspnet.login: invalid credentials")
		}

		// Generate JWT Token
		claims := jwt.MapClaims{
			"sub":      coerce.ToString(dbID),
			"username": dbUsername,
			"email":    dbEmail,
			"exp":      time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
			"iat":      time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return fmt.Errorf("aspnet.login: failed to sign token: %v", err)
		}

		// Save results to scope
		scope.Set(targetToken, tokenString)

		userMap := map[string]interface{}{
			"id":       coerce.ToString(dbID),
			"username": dbUsername,
			"email":    dbEmail,
		}
		for i, cf := range customFields {
			// Convert scanner interface value appropriately
			if valBytes, ok := customValues[i].([]byte); ok {
				userMap[cf] = string(valBytes)
			} else {
				userMap[cf] = customValues[i]
			}
		}
		scope.Set(targetUser, userMap)

		return nil
	}, engine.SlotMeta{
		Description: "Authenticate user using legacy ASP.NET Core Identity AspNetUsers table schema and PBKDF2 hashing.",
		Example:     "aspnet.login:\n  username: $input_user\n  password: $input_pass\n  fields: ['TenantId', 'FullName']\n  as: $token\n  user_as: $user",
		Inputs: map[string]engine.InputMeta{
			"username":   {Description: "Username or Email address of the user", Required: true, Type: "string"},
			"password":   {Description: "Plain-text password", Required: true, Type: "string"},
			"db":         {Description: "Database connection name (Default: 'default')", Required: false, Type: "string"},
			"secret":     {Description: "JWT secret key for signing", Required: false, Type: "string"},
			"expires_in": {Description: "Token expiration time in seconds (Default: 86400)", Required: false, Type: "int"},
			"fields":     {Description: "List of custom database columns to retrieve from AspNetUsers", Required: false, Type: "list"},
			"as":         {Description: "Variable to store the JWT token (Default: 'token')", Required: false, Type: "string"},
			"user_as":    {Description: "Variable to store the user data map (Default: 'user')", Required: false, Type: "string"},
		},
	})

	// 2. ASPNET.HASH
	eng.Register("aspnet.hash", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		input := coerce.ToString(resolveValue(node.Value, scope))
		iterations := 10000
		target := "hash_result"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "password" || c.Name == "text" || c.Name == "val" {
				input = coerce.ToString(val)
			}
			if c.Name == "iterations" || c.Name == "iter" {
				iterVal, _ := coerce.ToInt64(val)
				iterations = int(iterVal)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if input == "" {
			return fmt.Errorf("aspnet.hash: password is required")
		}

		hashed, err := HashAspNetHash(input, iterations)
		if err != nil {
			return fmt.Errorf("aspnet.hash: failed to hash: %v", err)
		}

		scope.Set(target, hashed)
		return nil
	}, engine.SlotMeta{
		Description: "Hash a plain-text password using the legacy ASP.NET Identity V3 PBKDF2/HMAC-SHA256 format.",
		Example:     "aspnet.hash: $input_pass\n  iterations: 10000\n  as: $db_hash",
		Inputs: map[string]engine.InputMeta{
			"(value)":    {Description: "The plain-text password to hash", Required: false, Type: "string"},
			"password":   {Description: "The plain-text password to hash", Required: false, Type: "string"},
			"iterations": {Description: "Iteration count (Default: 10000)", Required: false, Type: "int"},
			"as":         {Description: "Variable name to store the generated hash result (Default: 'hash_result')", Required: false, Type: "string"},
		},
	})

	// 3. ASPNET.VERIFY
	eng.Register("aspnet.verify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
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
			return fmt.Errorf("aspnet.verify: hash and password are required")
		}

		isValid := VerifyAspNetHash(hash, password)
		scope.Set(target, isValid)
		return nil
	}, engine.SlotMeta{
		Description: "Verify a plain-text password against an ASP.NET Identity V3 PBKDF2/HMAC-SHA256 hash.",
		Example:     "aspnet.verify\n  hash: $db_hash\n  password: $input_pass\n  as: $is_valid",
		Inputs: map[string]engine.InputMeta{
			"hash":     {Description: "The ASP.NET Identity V3 hash string (base64 encoded)", Required: true, Type: "string"},
			"password": {Description: "The plain-text password to verify", Required: true, Type: "string"},
			"as":       {Description: "Variable name to store the boolean result (Default: 'verify_result')", Required: false, Type: "string"},
		},
	})

	// 4. ASPNET.VALIDATE_PASSWORD
	eng.Register("aspnet.validate_password", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var password string
		requireDigit := true
		requireLowercase := true
		requireUppercase := true
		requireNonAlphanumeric := true
		requiredLength := 6
		requiredUniqueChars := 1
		target := "is_valid"
		targetErrors := "password_errors"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "password" || c.Name == "val" || c.Name == "value" {
				password = coerce.ToString(val)
			}
			if c.Name == "require_digit" {
				requireDigit, _ = coerce.ToBool(val)
			}
			if c.Name == "require_lowercase" {
				requireLowercase, _ = coerce.ToBool(val)
			}
			if c.Name == "require_uppercase" {
				requireUppercase, _ = coerce.ToBool(val)
			}
			if c.Name == "require_non_alphanumeric" {
				requireNonAlphanumeric, _ = coerce.ToBool(val)
			}
			if c.Name == "required_length" {
				valInt, _ := coerce.ToInt64(val)
				requiredLength = int(valInt)
			}
			if c.Name == "required_unique_chars" {
				valInt, _ := coerce.ToInt64(val)
				requiredUniqueChars = int(valInt)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "errors_as" {
				targetErrors = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if password == "" {
			scope.Set(target, false)
			scope.Set(targetErrors, []interface{}{"Password is required"})
			return nil
		}

		var errors []interface{}

		// 1. Length Check
		if len(password) < requiredLength {
			errors = append(errors, fmt.Sprintf("Passwords must be at least %d characters.", requiredLength))
		}

		// 2. Digit Check
		if requireDigit {
			hasDigit := false
			for _, r := range password {
				if r >= '0' && r <= '9' {
					hasDigit = true
					break
				}
			}
			if !hasDigit {
				errors = append(errors, "Passwords must have at least one digit ('0'-'9').")
			}
		}

		// 3. Lowercase Check
		if requireLowercase {
			hasLower := false
			for _, r := range password {
				if r >= 'a' && r <= 'z' {
					hasLower = true
					break
				}
			}
			if !hasLower {
				errors = append(errors, "Passwords must have at least one lowercase ('a'-'z').")
			}
		}

		// 4. Uppercase Check
		if requireUppercase {
			hasUpper := false
			for _, r := range password {
				if r >= 'A' && r <= 'Z' {
					hasUpper = true
					break
				}
			}
			if !hasUpper {
				errors = append(errors, "Passwords must have at least one uppercase ('A'-'Z').")
			}
		}

		// 5. Non-alphanumeric Check
		if requireNonAlphanumeric {
			hasNonAlpha := false
			for _, r := range password {
				isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
				isDigit := r >= '0' && r <= '9'
				if !isLetter && !isDigit {
					hasNonAlpha = true
					break
				}
			}
			if !hasNonAlpha {
				errors = append(errors, "Passwords must have at least one non alphanumeric character.")
			}
		}

		// 6. Unique Chars Check
		if requiredUniqueChars > 1 {
			uniqueSet := make(map[rune]bool)
			for _, r := range password {
				uniqueSet[r] = true
			}
			if len(uniqueSet) < requiredUniqueChars {
				errors = append(errors, fmt.Sprintf("Passwords must use at least %d different characters.", requiredUniqueChars))
			}
		}

		isValid := len(errors) == 0
		scope.Set(target, isValid)
		if !isValid {
			scope.Set(targetErrors, errors)
		} else {
			scope.Set(targetErrors, nil)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Validate a password against configurable ASP.NET Core Identity password policies.",
		Example:     "aspnet.validate_password: $password\n  required_length: 8\n  require_uppercase: true\n  errors_as: $pw_errors",
		Inputs: map[string]engine.InputMeta{
			"(value)":                  {Description: "The plain-text password to validate", Required: false, Type: "string"},
			"password":                 {Description: "The plain-text password to validate", Required: false, Type: "string"},
			"require_digit":            {Description: "Require at least one digit (Default: true)", Required: false, Type: "bool"},
			"require_lowercase":        {Description: "Require at least one lowercase letter (Default: true)", Required: false, Type: "bool"},
			"require_uppercase":        {Description: "Require at least one uppercase letter (Default: true)", Required: false, Type: "bool"},
			"require_non_alphanumeric": {Description: "Require at least one special character (Default: true)", Required: false, Type: "bool"},
			"required_length":          {Description: "Minimum length of the password (Default: 6)", Required: false, Type: "int"},
			"required_unique_chars":    {Description: "Minimum number of unique characters (Default: 1)", Required: false, Type: "int"},
			"as":                       {Description: "Variable name to store validation status (Default: 'is_valid')", Required: false, Type: "string"},
			"errors_as":                {Description: "Variable name to store list of error messages (Default: 'password_errors')", Required: false, Type: "string"},
		},
	})
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

// HashAspNetHash generates ASP.NET Identity V3 password hashes
// Format: 0x01 (1 byte) | Prf (4 bytes, SHA256=1) | IterCount (4 bytes) | SaltLen (4 bytes) | Salt | Subkey
func HashAspNetHash(password string, iterations int) (string, error) {
	if iterations <= 0 {
		iterations = 10000
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	subkey := pbkdf2.Key([]byte(password), salt, iterations, 32, sha256.New)

	blob := make([]byte, 1+4+4+4+len(salt)+len(subkey))
	blob[0] = 0x01
	binary.BigEndian.PutUint32(blob[1:5], 1) // PRF = SHA256
	binary.BigEndian.PutUint32(blob[5:9], uint32(iterations))
	binary.BigEndian.PutUint32(blob[9:13], uint32(len(salt)))
	copy(blob[13:29], salt)
	copy(blob[29:], subkey)

	return base64.StdEncoding.EncodeToString(blob), nil
}
