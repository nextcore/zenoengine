package worker

import "context"

// JobQueue defines the interface for a background job queue
type JobQueue interface {
	// Push adds a job to the queue
	Push(ctx context.Context, queue string, payload []byte) error

	// Pop retrieves a job from one of the specified queues (blocking or polling)
	// Returns queue name, payload, and error
	Pop(ctx context.Context, queues []string) (string, []byte, error)

	// Close cleans up resources
	Close() error
}
