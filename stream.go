package pedantigo

import (
	"encoding/json"
	"sync"
)

// StreamParser provides stateful parsing for streaming JSON chunks.
// Designed for LLM streaming APIs (Anthropic, OpenAI, etc.)
// Does NOT perform JSON repair - waits for complete valid JSON.
type StreamParser[T any] struct {
	validator *Validator[T]
	buffer    []byte
	mu        sync.Mutex
	attempts  int
}

// NewStreamParser creates a parser for streaming JSON.
func NewStreamParser[T any](opts ...ValidatorOptions) *StreamParser[T] {
	return &StreamParser[T]{
		validator: New[T](opts...),
		buffer:    make([]byte, 0),
	}
}

// NewStreamParserWithValidator creates a parser with a custom validator.
// Use this for discriminated unions or when you need custom validator options.
func NewStreamParserWithValidator[T any](validator *Validator[T]) *StreamParser[T] {
	return &StreamParser[T]{
		validator: validator,
		buffer:    make([]byte, 0),
	}
}

// Feed adds a new chunk of JSON data and returns the current state.
// Returns:
//   - *T: Parsed struct (nil if JSON incomplete)
//   - *StreamState: Completion state with tracking info
//   - error: Validation errors (only when complete), or nil
func (sp *StreamParser[T]) Feed(chunk []byte) (*T, *StreamState, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.buffer = append(sp.buffer, chunk...)
	sp.attempts++

	state := &StreamState{
		BytesReceived: len(sp.buffer),
		ParseAttempts: sp.attempts,
	}

	// First, check if JSON is complete using standard unmarshal
	var temp T
	if err := json.Unmarshal(sp.buffer, &temp); err != nil {
		state.LastError = err
		state.IsComplete = false
		return nil, state, nil // Not an error, just incomplete
	}

	// JSON is complete - use Pedantigo's Unmarshal for proper required field detection
	state.IsComplete = true

	obj, err := sp.validator.Unmarshal(sp.buffer)
	if err != nil {
		return obj, state, err // Return struct with validation errors
	}

	return obj, state, nil
}

// Reset clears the buffer and starts fresh.
func (sp *StreamParser[T]) Reset() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.buffer = make([]byte, 0)
	sp.attempts = 0
}

// Buffer returns the current accumulated buffer (for debugging).
func (sp *StreamParser[T]) Buffer() []byte {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	result := make([]byte, len(sp.buffer))
	copy(result, sp.buffer)
	return result
}
