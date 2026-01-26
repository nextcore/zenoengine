package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"zeno/pkg/worker"
)

func RegisterJobSlots(eng *engine.Engine, queue worker.JobQueue, setConfig func([]string)) {

	// WORKER.CONFIG
	eng.Register("worker.config", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		if setConfig == nil {
			return nil
		}
		var queues []string
		if node.Value != nil {
			val := coerce.ToString(node.Value)
			if val != "" {
				parts := strings.Split(val, ",")
				for _, p := range parts {
					queues = append(queues, strings.TrimSpace(p))
				}
			}
		}

		// Handle array children if present or comma separated string?
		// Let's support simple array of strings as children or single string value.
		for _, c := range node.Children {
			v := coerce.ToString(parseNodeValue(c, scope))
			if v == "" && c.Name != "" {
				v = c.Name
			}
			if v != "" {
				queues = append(queues, v)
			}
		}

		if len(queues) > 0 {
			setConfig(queues)
		}
		return nil
	}, engine.SlotMeta{
		Description: "Configure worker queues.",
		Example: `worker.config
  - "high_priority"
  - "default"`,
	})

	// JOB.ENQUEUE
	eng.Register("job.enqueue", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// [SAFETY] Cek Queue
		if queue == nil {
			return fmt.Errorf("job.enqueue failed: Queue is not available")
		}

		var queueName string = "default"
		var payload interface{}

		// Support shorthand: job.enqueue: "email_queue"
		if node.Value != nil && fmt.Sprintf("%v", node.Value) != "" {
			queueName = coerce.ToString(node.Value)
		}

		for _, c := range node.Children {
			if c.Name == "queue" {
				queueName = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "payload" {
				// Payload bisa map kompleks, gunakan parseNodeValue
				payload = parseNodeValue(c, scope)
			}
		}

		if payload == nil {
			return fmt.Errorf("job.enqueue: payload is required")
		}

		// Marshal payload ke JSON string untuk disimpan di Redis List
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("job.enqueue: failed to marshal payload: %v", err)
		}

		// Push ke Queue
		err = queue.Push(ctx, queueName, jsonBytes)
		if err != nil {
			return err
		}

		return nil
	}, engine.SlotMeta{
		Description: "Add a job to the background queue (Redis/DB).",
		Example: `job.enqueue
  queue: "emails"
  payload:
    to: "budi@example.com"
    subject: "Welcome"`,
	})
}
