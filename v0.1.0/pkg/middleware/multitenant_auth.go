package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// MultiTenantAuth is a Chi middleware that handles multi-tenant authentication
// It auto-detects tenant from X-Tenant-ID header or subdomain,
// validates JWT token, and sets auth context for downstream handlers
func MultiTenantAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Auto-detect tenant ID (Header priority -> Subdomain fallback)
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				// Fallback to subdomain
				host := r.Host
				parts := strings.Split(host, ".")
				if len(parts) >= 2 && parts[0] != "localhost" && parts[0] != "127" {
					tenantID = parts[0]
				}
			}

			// 2. Validate tenant from system database if detected
			if tenantID != "" {
				dbVal := r.Context().Value("database")
				if dbVal != nil {
					db := dbVal.(*sql.DB)

					var dbConnectionName, tenantName, tenantCode string
					var isActive bool
					query := "SELECT code, name, db_connection_name, is_active FROM tenants WHERE code = ? AND is_active = 1 LIMIT 1"
					err := db.QueryRowContext(r.Context(), query, tenantID).Scan(&tenantCode, &tenantName, &dbConnectionName, &isActive)

					if err == nil && isActive {
						// Store tenant info in context
						ctx := r.Context()
						ctx = context.WithValue(ctx, "CURRENT_TENANT_DB", dbConnectionName)
						ctx = context.WithValue(ctx, "CURRENT_TENANT_NAME", tenantName)
						ctx = context.WithValue(ctx, "CURRENT_TENANT_ID", tenantCode)
						r = r.WithContext(ctx)
					}
					// If tenant validation fails, continue anyway (for single-tenant apps)
				}
			}

			// 3. Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Support Cookie fallback
				if cookie, err := r.Cookie("token"); err == nil {
					authHeader = "Bearer " + cookie.Value
				}
			}

			if authHeader == "" {
				http.Error(w, `{"error":"UNAUTHORIZED","message":"Authentication required","success":false}`, http.StatusUnauthorized)
				return
			}

			// 4. Verify JWT token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"UNAUTHORIZED","message":"Invalid or expired token","success":false}`, http.StatusUnauthorized)
				return
			}

			// 5. Extract claims and store in context
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx := r.Context()

				// Store full claims as "session" (for backward compatibility)
				ctx = context.WithValue(ctx, "session", claims)

				// Store auth object (for POS API compatibility)
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
				ctx = context.WithValue(ctx, "auth", authObj)

				r = r.WithContext(ctx)
			}

			// 6. Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// InjectAuthToScope injects auth data from HTTP context into ZenoLang scope
// This should be called in createHandler before executing ZenoLang code
func InjectAuthToScope(r *http.Request, scope interface{}) {
	// This will be implemented to bridge HTTP context to ZenoLang scope
	// For now, it's a placeholder that the engine will use

	type Scope interface {
		Set(key string, value interface{})
	}

	if s, ok := scope.(Scope); ok {
		// Inject session
		if session := r.Context().Value("session"); session != nil {
			s.Set("session", session)
		}

		// Inject auth object
		if auth := r.Context().Value("auth"); auth != nil {
			s.Set("auth", auth)
		}

		// Inject tenant info
		if tenantDB := r.Context().Value("CURRENT_TENANT_DB"); tenantDB != nil {
			s.Set("CURRENT_TENANT_DB", tenantDB)
		}
		if tenantName := r.Context().Value("CURRENT_TENANT_NAME"); tenantName != nil {
			s.Set("CURRENT_TENANT_NAME", tenantName)
		}
		if tenantID := r.Context().Value("CURRENT_TENANT_ID"); tenantID != nil {
			s.Set("CURRENT_TENANT_ID", tenantID)
		}
	}
}
