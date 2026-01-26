package slots

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func RegisterAuthSlots(eng *engine.Engine, dbMgr *dbmanager.DBManager) {

	// 1. AUTH.LOGIN
	eng.Register("auth.login", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var username, password string
		table := "users"
		colUser := "email"
		colPass := "password"
		jwtSecret := "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar" // Default from .env
		target := "token"
		dbName := "default"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "username" || c.Name == "email" {
				username = coerce.ToString(val)
			}
			if c.Name == "password" {
				password = coerce.ToString(val)
			}
			if c.Name == "table" {
				table = coerce.ToString(val)
			}
			if c.Name == "col_user" {
				colUser = coerce.ToString(val)
			}
			if c.Name == "col_pass" {
				colPass = coerce.ToString(val)
			}
			if c.Name == "secret" {
				jwtSecret = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "db" {
				dbName = coerce.ToString(val)
			}
		}

		if username == "" || password == "" {
			return fmt.Errorf("auth.login: username and password required")
		}

		// DB Lookup
		db := dbMgr.GetConnection(dbName)
		dialect := dbMgr.GetDialect(dbName)
		if db == nil {
			return fmt.Errorf("auth.login: database connection '%s' not found", dbName)
		}

		query := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s WHERE %s = %s%s",
			dialect.QuoteIdentifier("id"),
			dialect.QuoteIdentifier("username"),
			dialect.QuoteIdentifier(colUser),
			dialect.QuoteIdentifier(colPass),
			dialect.QuoteIdentifier(table),
			dialect.QuoteIdentifier(colUser),
			dialect.Placeholder(1),
			dialect.Limit(1, 0))

		fmt.Printf("[AUTH DEBUG] Query: %s\n", query)
		fmt.Printf("[AUTH DEBUG] Username: %s, Table: %s, DB: %s\n", username, table, dbName)

		var id int
		var dbUsername, dbUser, dbPass string

		err := db.QueryRowContext(ctx, query, username).Scan(&id, &dbUsername, &dbUser, &dbPass)
		if err != nil {
			fmt.Printf("[AUTH DEBUG] DB Query Error: %v\n", err)
			return fmt.Errorf("auth.login: invalid credentials")
		}

		fmt.Printf("[AUTH DEBUG] User found: ID=%d, Username=%s, Email=%s\n", id, dbUsername, dbUser)

		// Verify Password
		if err := bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(password)); err != nil {
			fmt.Printf("[AUTH DEBUG] Password mismatch: %v\n", err)
			return fmt.Errorf("auth.login: invalid credentials")
		}

		fmt.Printf("[AUTH DEBUG] Password verified successfully\n")

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  id,
			"username": dbUsername,
			"email":    dbUser,
			"exp":      time.Now().Add(time.Hour * 72).Unix(),
		})

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return err
		}

		scope.Set(target, tokenString)
		return nil
	}, engine.SlotMeta{
		Description: "Verify user credentials and return a JWT token.",
		Example:     "auth.login\n  username: $user\n  password: $pass\n  as: $token",
		Inputs: map[string]engine.InputMeta{
			"username": {Description: "Email or Username", Required: false},
			"email":    {Description: "Alias for username", Required: false},
			"password": {Description: "Password", Required: true},
			"table":    {Description: "User table name (Default: 'users')", Required: false},
			"col_user": {Description: "Email/Username column (Default: 'email')", Required: false},
			"col_pass": {Description: "Password column (Default: 'password')", Required: false},
			"secret":   {Description: "JWT Secret key", Required: false},
			"db":       {Description: "Database connection name (Default: 'default')", Required: false},
			"as":       {Description: "Variable to store token", Required: false},
		},
	})

	// 2. AUTH.MIDDLEWARE (Guard) - Auto Multi-Tenant Detection
	eng.Register("auth.middleware", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		jwtSecret := "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar" // Default from .env
		var doNode *engine.Node

		// Parse parameters
		for _, c := range node.Children {
			if c.Name == "secret" {
				jwtSecret = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "do" {
				doNode = c
			}
		}

		// Get HTTP request
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return fmt.Errorf("auth.middleware: not in http context")
		}
		r := reqVal.(*http.Request)

		// [AUTO] Multi-Tenant Detection (ALWAYS enabled for POS API compatibility)
		// Check X-Tenant-ID header first, fallback to subdomain
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			// Fallback to subdomain
			host := r.Host
			parts := strings.Split(host, ".")
			if len(parts) >= 2 && parts[0] != "localhost" && parts[0] != "127" {
				tenantID = parts[0]
			}
		}

		// [AUTO] Validate tenant from system database if tenant detected
		if tenantID != "" {
			dbVal := ctx.Value("database")
			if dbVal != nil {
				db := dbVal.(*sql.DB)

				var dbConnectionName, tenantName, tenantCode string
				var isActive bool
				query := "SELECT code, name, db_connection_name, is_active FROM tenants WHERE code = ? AND is_active = 1 LIMIT 1"
				err := db.QueryRowContext(ctx, query, tenantID).Scan(&tenantCode, &tenantName, &dbConnectionName, &isActive)

				if err == nil && isActive {
					// Set tenant info in scope (for POS API)
					scope.Set("CURRENT_TENANT_DB", dbConnectionName)
					scope.Set("CURRENT_TENANT_NAME", tenantName)
					scope.Set("CURRENT_TENANT_ID", tenantCode)
				}
				// If tenant validation fails, continue anyway (for single-tenant apps)
			}
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			// Support Cookie fallback (check 'token' or 'auth_token')
			if cookie, err := r.Cookie("token"); err == nil {
				authHeader = "Bearer " + cookie.Value
			} else if cookie, err := r.Cookie("auth_token"); err == nil {
				authHeader = "Bearer " + cookie.Value
			}
		}

		// Check Redirect
		redirectURL := ""
		for _, c := range node.Children {
			if c.Name == "redirect" {
				redirectURL = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		if authHeader == "" {
			if redirectURL != "" {
				wVal := ctx.Value("httpWriter")
				if w, ok := wVal.(http.ResponseWriter); ok {
					http.Redirect(w, r, redirectURL, http.StatusFound)
					return ErrReturn // Stop execution
				}
			}
			return fmt.Errorf("unauthorized: missing token")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			if redirectURL != "" {
				wVal := ctx.Value("httpWriter")
				if w, ok := wVal.(http.ResponseWriter); ok {
					http.Redirect(w, r, redirectURL, http.StatusFound)
					return ErrReturn // Stop execution
				}
			}
			return fmt.Errorf("unauthorized: invalid token")
		}

		// Set User to Scope
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			scope.Set("session", claims) // Existing behavior for backward compatibility

			// [AUTO] ALWAYS set $auth object for POS API compatibility
			authObj := make(map[string]interface{})
			if userID, ok := claims["user_id"]; ok {
				authObj["user_id"] = userID
			}
			if email, ok := claims["email"]; ok {
				authObj["email"] = email
			}
			if tid, ok := claims["tenant_id"]; ok {
				authObj["tenant_id"] = tid
			}
			if role, ok := claims["role"]; ok {
				authObj["role"] = role
			}
			scope.Set("auth", authObj)
		}

		// Exec 'do' block (Protected Routes)
		if doNode != nil {
			for _, child := range doNode.Children {
				if err := eng.Execute(ctx, child, scope); err != nil {
					return err
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Protect routes with JWT verification. Supports multi-tenant with subdomain detection.",
		Example:     "auth.middleware {\n  do: {\n     log: 'Hello Admin'\n  }\n}\n\n// Multi-tenant:\nauth.middleware {\n  tenant_header: \"X-Tenant-ID\"\n  tenant_db_lookup: true\n  set_auth_object: true\n  do: { ... }\n}",
		Inputs: map[string]engine.InputMeta{
			"secret":           {Description: "JWT Secret key", Required: false},
			"redirect":         {Description: "Login URL for redirect on failure", Required: false},
			"tenant_header":    {Description: "Header name for tenant ID (fallback to subdomain)", Required: false},
			"tenant_db_lookup": {Description: "Enable tenant validation from system DB", Required: false},
			"set_auth_object":  {Description: "Set $auth object with user_id, email, etc", Required: false},
		},
	})

	// 3. AUTH.USER
	eng.Register("auth.user", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "user"

		// auth.user: $current_user
		if node.Value != nil {
			target = strings.TrimPrefix(coerce.ToString(node.Value), "$")
		}

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if val, ok := scope.Get("session"); ok {
			scope.Set(target, val)
			return nil
		}

		// Fallback: Check cookies and decode JWT
		reqVal := ctx.Value("httpRequest")
		if req, ok := reqVal.(*http.Request); ok {
			var tokenString string
			if cookie, err := req.Cookie("auth_token"); err == nil {
				tokenString = cookie.Value
			} else if cookie, err := req.Cookie("token"); err == nil {
				tokenString = cookie.Value
			}

			if tokenString != "" {
				// We need the secret. Since it's not passed here, we use the default
				// or we could try to find it from env.
				jwtSecret := os.Getenv("JWT_SECRET")
				if jwtSecret == "" {
					jwtSecret = "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar"
				}

				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})

				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						scope.Set(target, claims)
						return nil
					}
				}
			}
		}

		scope.Set(target, nil)
		return nil
	}, engine.SlotMeta{
		Description: "Retrieve user data from current session.",
		Example:     "auth.user: $current_user",
		Inputs: map[string]engine.InputMeta{
			"as": {Description: "Variable to store user data", Required: false},
		},
	})

	// 4. JWT.SIGN
	eng.Register("jwt.sign", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var secret string
		expiresIn := int64(86400) // 24 hours default
		target := "token"
		claims := make(map[string]interface{})

		for _, c := range node.Children {
			if c.Name == "secret" {
				secret = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "expires_in" || c.Name == "expiry" {
				expiresIn, _ = coerce.ToInt64(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "claims" {
				// Parse claims as map
				if claimsVal := parseNodeValue(c, scope); claimsVal != nil {
					if claimsMap, ok := claimsVal.(map[string]interface{}); ok {
						claims = claimsMap
					}
				}
			}
		}

		if secret == "" {
			return fmt.Errorf("jwt.sign: secret is required")
		}

		// Add expiry to claims
		claims["exp"] = time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
		claims["iat"] = time.Now().Unix()

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			return fmt.Errorf("jwt.sign: failed to sign token: %v", err)
		}

		scope.Set(target, tokenString)
		return nil
	}, engine.SlotMeta{
		Description: "Generate JWT token with custom claims",
		Example:     "jwt.sign:\n  secret: env(\"JWT_SECRET\")\n  claims: { user_id: $user.id }\n  expires_in: 86400\n  as: $token",
		Inputs: map[string]engine.InputMeta{
			"secret":     {Description: "JWT Secret key", Required: true},
			"claims":     {Description: "Token claims as map", Required: true},
			"expires_in": {Description: "Expiry in seconds (default: 86400)", Required: false},
			"as":         {Description: "Variable to store token", Required: false},
		},
	})

	// 5. JWT.VERIFY
	eng.Register("jwt.verify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var tokenString, secret string
		target := "claims"

		// Support direct value: jwt.verify: $token
		if node.Value != nil {
			tokenString = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "token" || c.Name == "val" {
				tokenString = coerce.ToString(val)
			}
			if c.Name == "secret" {
				secret = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// Jika secret kosong, cari dari environment global atau default
		if secret == "" {
			secret = "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar" // Default from .env
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return fmt.Errorf("jwt.verify: invalid token or expired")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			scope.Set(target, claims)
		} else {
			return fmt.Errorf("jwt.verify: invalid claims structure")
		}

		return nil
	}, engine.SlotMeta{
		Description: "Explicitly verify a JWT token and retrieve its claims.",
		Example:     "jwt.verify: $token\n  secret: 'shhh'\n  as: $user_data",
		Inputs: map[string]engine.InputMeta{
			"token":  {Description: "Token String", Required: false},
			"secret": {Description: "Secret Key", Required: false},
			"as":     {Description: "Resulting claims", Required: false},
		},
	})

	// 5. JWT.REFRESH
	eng.Register("jwt.refresh", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Disclaimer: Simple refresh (create new token with new expiry based on old claims)
		// NOT implementing refresh token rotation with DB lookup in this slot for simplicity.
		var tokenString, secret string
		expirySeconds := int64(86400) // 24 Hours default
		target := "new_token"

		if node.Value != nil {
			tokenString = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "token" {
				tokenString = coerce.ToString(val)
			}
			if c.Name == "secret" {
				secret = coerce.ToString(val)
			}
			if c.Name == "expiry" {
				expirySeconds, _ = coerce.ToInt64(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if secret == "" {
			secret = "rahasia_dapur_pekalongan_kota_2025_!@#_jgn_disebar" // Default from .env
		}

		// Parse token (even if expired, we might want to allow refresh if within grace period - but here we require valid signature)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		// Note: jwt.Parse usually errors if expired.
		// For refresh flow, we might need ParseUnverified or custom checking.
		// Here we assume client sends a STILL VALID token to get a new one (short lived rotation).

		if err != nil {
			// Try parsing without validation to extract claims if it's just expired
			// But for security, let's stick to requiring a valid token or handle error externally.
			return fmt.Errorf("jwt.refresh: invalid source token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Update Exp
			claims["exp"] = time.Now().Add(time.Duration(expirySeconds) * time.Second).Unix()

			newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedStr, err := newToken.SignedString([]byte(secret))
			if err != nil {
				return err
			}
			scope.Set(target, signedStr)
		}

		return nil
	}, engine.SlotMeta{
		Description: "Refresh JWT token with a new expiration.",
		Example:     "jwt.refresh: $old_token\n  as: $new_token",
	})
}
