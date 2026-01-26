package slots

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"zeno/pkg/engine"
)

func TestNativeAuthJson(t *testing.T) {
	// 1. Setup Engine
	eng := engine.NewEngine()
	RegisterBladeSlots(eng) // Native
	
	// 2. Mock View
	os.Mkdir("views", 0755)
	defer os.RemoveAll("views")
	
	content := `
@auth
  LOGGED_IN_USER
@endauth

@guest
  GUEST_USER
@endguest

<script>
var data = @json($data);
</script>
`
	os.WriteFile("views/test_auth.blade.zl", []byte(content), 0644)
	
	// 3. Execution - SCENARIO 1: Guest (no user var)
	root := &engine.Node{ Name: "view.blade", Value: "test_auth.blade.zl" }
	scopeGuest := engine.NewScope(nil)
	scopeGuest.Set("data", map[string]interface{}{"a": 1, "b": "foo"})
	
	recGuest := httptest.NewRecorder()
	ctxGuest := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(recGuest))
	
	if err := eng.Execute(ctxGuest, root, scopeGuest); err != nil {
		t.Fatalf("Guest Exec failed: %v", err)
	}
	
	outGuest := strings.TrimSpace(recGuest.Body.String())
	if strings.Contains(outGuest, "LOGGED_IN_USER") { t.Errorf("Guest should not see LOGGED_IN_USER") }
	if !strings.Contains(outGuest, "GUEST_USER") { t.Errorf("Guest should see GUEST_USER") }
	if !strings.Contains(outGuest, `{"a":1,"b":"foo"}`) { t.Errorf("JSON output mismatch: %s", outGuest) }


	// 3. Execution - SCENARIO 2: Auth (user var set)
	scopeAuth := engine.NewScope(nil)
	scopeAuth.Set("user", map[string]string{"name": "Admin"}) // "user" variable implies logged in
	scopeAuth.Set("data", map[string]interface{}{"a": 1})

	recAuth := httptest.NewRecorder()
	ctxAuth := context.WithValue(context.Background(), "httpWriter", http.ResponseWriter(recAuth))
	
	if err := eng.Execute(ctxAuth, root, scopeAuth); err != nil {
		t.Fatalf("Auth Exec failed: %v", err)
	}
	
	outAuth := strings.TrimSpace(recAuth.Body.String())
	if !strings.Contains(outAuth, "LOGGED_IN_USER") { t.Errorf("Auth should see LOGGED_IN_USER") }
	if strings.Contains(outAuth, "GUEST_USER") { t.Errorf("Auth should not see GUEST_USER") }
}
