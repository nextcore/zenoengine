package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"zeno/internal/slots"
	"zeno/pkg/engine"
	pkgslots "zeno/pkg/slots"
)

type ExecuteRequest struct {
	Script string `json:"script"`
}

type ExecuteResponse struct {
	Success   bool                   `json:"success"`
	Error     *string                `json:"error"`
	Variables map[string]interface{} `json:"variables"`
}

func main() {
	eng := engine.NewEngine()
	// Register equivalent slots
	slots.RegisterMathSlots(eng)
	slots.RegisterTimeSlots(eng)
	pkgslots.RegisterLogicSlots(eng)
	slots.RegisterFunctionSlots(eng)
	slots.RegisterCollectionSlots(eng)
	slots.RegisterJSONSlots(eng)

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var payload ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			errStr := err.Error()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExecuteResponse{
				Success:   false,
				Error:     &errStr,
				Variables: make(map[string]interface{}),
			})
			return
		}

		root, err := engine.ParseString(payload.Script, "request.zl")
		if err != nil {
			errStr := fmt.Sprintf("Parsing error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExecuteResponse{
				Success:   false,
				Error:     &errStr,
				Variables: make(map[string]interface{}),
			})
			return
		}

		scope := engine.NewScope(nil)
		ctx := context.Background()

		if err := eng.Execute(ctx, root, scope); err != nil {
			errStr := fmt.Sprintf("Execution error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExecuteResponse{
				Success:   false,
				Error:     &errStr,
				Variables: make(map[string]interface{}),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ExecuteResponse{
			Success:   true,
			Error:     nil,
			Variables: scope.ToMap(),
		})
	})

	fmt.Println("🚀 Go ZenoEngine benchmark server running at http://127.0.0.1:4000")
	if err := http.ListenAndServe("127.0.0.1:4000", nil); err != nil {
		panic(err)
	}
}
