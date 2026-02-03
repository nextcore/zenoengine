package slots

import (
	"context"
	"encoding/json"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

type MockJobQueue struct {
	LastQueue   string
	LastPayload []byte
}

func (m *MockJobQueue) Push(ctx context.Context, queue string, payload []byte) error {
	m.LastQueue = queue
	m.LastPayload = payload
	return nil
}

func (m *MockJobQueue) Pop(ctx context.Context, queues []string) (string, []byte, error) {
	return "", nil, nil
}

func (m *MockJobQueue) Close() error {
	return nil
}

func TestJobSlots(t *testing.T) {
	eng := engine.NewEngine()
	mockQueue := &MockJobQueue{}
	RegisterJobSlots(eng, mockQueue, nil)

	t.Run("job.enqueue", func(t *testing.T) {
		scope := engine.NewScope(nil)
		payload := map[string]interface{}{"task": "email", "to": "a@b.c"}
		scope.Set("data", payload)

		node := &engine.Node{
			Name: "job.enqueue",
			Children: []*engine.Node{
				{Name: "queue", Value: "high"},
				{Name: "payload", Value: "$data"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		assert.Equal(t, "high", mockQueue.LastQueue)

		var received map[string]interface{}
		json.Unmarshal(mockQueue.LastPayload, &received)
		assert.Equal(t, "email", received["task"])
	})
}
