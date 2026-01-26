package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterJSONSlots(eng *engine.Engine) {

	// 1. JSON.PARSE
	eng.Register("json.parse", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var jsonStr string
		target := "json_result"

		if node.Value != nil {
			jsonStr = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "val" {
				jsonStr = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if jsonStr == "" {
			return fmt.Errorf("json.parse: input value is empty")
		}

		var result interface{}
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			return fmt.Errorf("json.parse: invalid json format: %v", err)
		}

		scope.Set(target, result)
		return nil
	}, engine.SlotMeta{Example: "json.parse: $response_body\n  as: $data"})

	// 2. JSON.STRINGIFY
	eng.Register("json.stringify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var input interface{}
		target := "json_string"

		if node.Value != nil {
			input = resolveValue(node.Value, scope)
		}

		for _, c := range node.Children {
			if c.Name == "val" {
				input = parseNodeValue(c, scope)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		bytes, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("json.stringify: failed to marshal: %v", err)
		}

		scope.Set(target, string(bytes))
		return nil
	}, engine.SlotMeta{Example: "json.stringify: $data\n  as: $json_str"})
}
