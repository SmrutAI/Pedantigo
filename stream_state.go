package pedantigo

// StreamState tracks the parsing progress of streaming JSON.
type StreamState struct {
	// IsComplete is true if JSON parsing succeeded
	IsComplete bool

	// BytesReceived is the total bytes accumulated
	BytesReceived int

	// ParseAttempts tracks how many times parsing was attempted
	ParseAttempts int

	// LastError holds the most recent parse error (nil if complete)
	LastError error

	// PresentFields lists JSON field paths that were successfully parsed
	// Only populated when IsComplete is true
	PresentFields []string
}

// HasField checks if a specific field path is present in the parsed result.
func (ss *StreamState) HasField(path string) bool {
	for _, f := range ss.PresentFields {
		if f == path {
			return true
		}
	}
	return false
}
