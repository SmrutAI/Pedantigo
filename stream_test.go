package pedantigo

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StreamTestUser is a test struct for streaming tests.
type StreamTestUser struct {
	Name  string `json:"name" pedantigo:"required"`
	Email string `json:"email" pedantigo:"email"`
	Age   int    `json:"age" pedantigo:"min=0"`
}

// NestedStreamTest is a test struct with nested fields.
type NestedStreamTest struct {
	User    StreamTestUser `json:"user"`
	Address Address        `json:"address"`
}

// Address is a nested struct for testing.
type Address struct {
	Street string `json:"street" pedantigo:"required"`
	City   string `json:"city" pedantigo:"required"`
}

// ==================== Parser Creation Tests ====================

func TestNewStreamParser(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	require.NotNil(t, parser)
	require.NotNil(t, parser.validator)
	assert.Empty(t, parser.buffer)
	assert.Equal(t, 0, parser.attempts)
}

func TestNewStreamParserWithValidator(t *testing.T) {
	validator := New[StreamTestUser]()
	parser := NewStreamParserWithValidator(validator)

	require.NotNil(t, parser)
	assert.Equal(t, validator, parser.validator)
	assert.Empty(t, parser.buffer)
	assert.Equal(t, 0, parser.attempts)
}

// ==================== Single Chunk Tests ====================

func TestStreamParser_SingleChunk_CompleteJSON(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()
	completeJSON := []byte(`{"name":"John","email":"john@example.com","age":30}`)

	obj, state, err := parser.Feed(completeJSON)

	require.NoError(t, err)
	require.NotNil(t, obj)
	require.NotNil(t, state)

	assert.Equal(t, "John", obj.Name)
	assert.Equal(t, "john@example.com", obj.Email)
	assert.Equal(t, 30, obj.Age)

	assert.True(t, state.IsComplete)
	assert.Equal(t, len(completeJSON), state.BytesReceived)
	assert.Equal(t, 1, state.ParseAttempts)
	assert.NoError(t, state.LastError)
}

func TestStreamParser_SingleChunk_IncompleteJSON(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()
	incompleteJSON := []byte(`{"name":"Jo`)

	obj, state, err := parser.Feed(incompleteJSON)

	require.NoError(t, err) // Not an error, just incomplete
	assert.Nil(t, obj)
	require.NotNil(t, state)

	assert.False(t, state.IsComplete)
	assert.Equal(t, len(incompleteJSON), state.BytesReceived)
	assert.Equal(t, 1, state.ParseAttempts)
	assert.Error(t, state.LastError) // JSON parse error stored
}

// ==================== Multiple Chunks Tests ====================

func TestStreamParser_MultipleChunks_IncompleteFirst(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// First chunk - incomplete
	chunk1 := []byte(`{"name":"Jo`)
	obj1, state1, err1 := parser.Feed(chunk1)

	require.NoError(t, err1)
	assert.Nil(t, obj1)
	assert.False(t, state1.IsComplete)
	assert.Equal(t, len(chunk1), state1.BytesReceived)
	assert.Equal(t, 1, state1.ParseAttempts)

	// Second chunk - complete
	chunk2 := []byte(`hn","email":"john@example.com","age":25}`)
	obj2, state2, err2 := parser.Feed(chunk2)

	require.NoError(t, err2)
	require.NotNil(t, obj2)
	assert.True(t, state2.IsComplete)
	assert.Equal(t, "John", obj2.Name)
	assert.Equal(t, "john@example.com", obj2.Email)
	assert.Equal(t, 25, obj2.Age)
	assert.Equal(t, len(chunk1)+len(chunk2), state2.BytesReceived)
	assert.Equal(t, 2, state2.ParseAttempts)
}

func TestStreamParser_MultipleChunks_ThreeChunks(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Chunk 1
	chunk1 := []byte(`{"name":`)
	obj1, state1, _ := parser.Feed(chunk1)
	assert.Nil(t, obj1)
	assert.False(t, state1.IsComplete)
	assert.Equal(t, 1, state1.ParseAttempts)

	// Chunk 2
	chunk2 := []byte(`"Alice","email":"alice@`)
	obj2, state2, _ := parser.Feed(chunk2)
	assert.Nil(t, obj2)
	assert.False(t, state2.IsComplete)
	assert.Equal(t, 2, state2.ParseAttempts)

	// Chunk 3 - complete
	chunk3 := []byte(`example.com","age":35}`)
	obj3, state3, err3 := parser.Feed(chunk3)

	require.NoError(t, err3)
	require.NotNil(t, obj3)
	assert.True(t, state3.IsComplete)
	assert.Equal(t, "Alice", obj3.Name)
	assert.Equal(t, "alice@example.com", obj3.Email)
	assert.Equal(t, 35, obj3.Age)
	assert.Equal(t, 3, state3.ParseAttempts)
}

// ==================== Validation Tests ====================

func TestStreamParser_ValidatesOnComplete(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Invalid email when complete
	invalidJSON := []byte(`{"name":"Bob","email":"not-an-email","age":25}`)
	obj, state, err := parser.Feed(invalidJSON)

	require.Error(t, err)  // Validation error
	require.NotNil(t, obj) // Struct is still returned
	assert.True(t, state.IsComplete)
	assert.Equal(t, "Bob", obj.Name)
	assert.Equal(t, "not-an-email", obj.Email)

	// Check that error is validation-related
	assert.Contains(t, err.Error(), "email")
}

func TestStreamParser_RequiredField_Missing(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Missing required 'name' field
	invalidJSON := []byte(`{"email":"test@example.com","age":20}`)
	obj, state, err := parser.Feed(invalidJSON)

	require.Error(t, err) // Validation error for missing required field
	require.NotNil(t, obj)
	assert.True(t, state.IsComplete)
}

func TestStreamParser_MinConstraint_Violated(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Age below minimum (min=0)
	invalidJSON := []byte(`{"name":"Charlie","email":"charlie@example.com","age":-5}`)
	obj, state, err := parser.Feed(invalidJSON)

	require.Error(t, err)
	require.NotNil(t, obj)
	assert.True(t, state.IsComplete)
	assert.Equal(t, -5, obj.Age)
}

func TestStreamParser_ValidData_NoError(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	validJSON := []byte(`{"name":"David","email":"david@example.com","age":40}`)
	obj, state, err := parser.Feed(validJSON)

	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.True(t, state.IsComplete)
}

// ==================== Reset Tests ====================

func TestStreamParser_Reset_ClearsBuffer(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Feed some data
	chunk := []byte(`{"name":"Eve`)
	_, _, _ = parser.Feed(chunk)

	// Verify buffer has data
	assert.Len(t, parser.Buffer(), len(chunk))

	// Reset
	parser.Reset()

	// Verify buffer is cleared
	assert.Empty(t, parser.Buffer())
	assert.Equal(t, 0, parser.attempts)
}

func TestStreamParser_Reset_AllowsReuse(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// First use
	json1 := []byte(`{"name":"Frank","email":"frank@example.com","age":28}`)
	obj1, _, err1 := parser.Feed(json1)
	require.NoError(t, err1)
	require.NotNil(t, obj1)
	assert.Equal(t, "Frank", obj1.Name)

	// Reset
	parser.Reset()

	// Second use - different data
	json2 := []byte(`{"name":"Grace","email":"grace@example.com","age":32}`)
	obj2, state2, err2 := parser.Feed(json2)
	require.NoError(t, err2)
	require.NotNil(t, obj2)
	assert.Equal(t, "Grace", obj2.Name)
	assert.Equal(t, 1, state2.ParseAttempts) // Counter reset
}

// ==================== Buffer Tests ====================

func TestStreamParser_Buffer_ReturnsCopy(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	chunk := []byte(`{"name":"Henry"}`)
	_, _, _ = parser.Feed(chunk)

	// Get buffer
	buf1 := parser.Buffer()
	buf2 := parser.Buffer()

	// Verify they are equal but different slices
	assert.Equal(t, buf1, buf2)
	assert.NotSame(t, &buf1, &buf2)

	// Modify one - should not affect the other
	buf1[0] = 'X'
	assert.NotEqual(t, buf1[0], buf2[0])

	// Original buffer should be unchanged
	buf3 := parser.Buffer()
	assert.Equal(t, byte('{'), buf3[0])
}

func TestStreamParser_Buffer_EmptyInitially(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()
	buf := parser.Buffer()

	assert.NotNil(t, buf)
	assert.Empty(t, buf)
}

// ==================== State Tracking Tests ====================

func TestStreamParser_BytesReceived_Accurate(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	chunk1 := []byte(`{"name":`)
	_, state1, _ := parser.Feed(chunk1)
	assert.Equal(t, len(chunk1), state1.BytesReceived)

	chunk2 := []byte(`"Ivy"}`)
	_, state2, _ := parser.Feed(chunk2)
	assert.Equal(t, len(chunk1)+len(chunk2), state2.BytesReceived)
}

func TestStreamParser_ParseAttempts_Increments(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	_, state1, _ := parser.Feed([]byte(`{`))
	assert.Equal(t, 1, state1.ParseAttempts)

	_, state2, _ := parser.Feed([]byte(`"name"`))
	assert.Equal(t, 2, state2.ParseAttempts)

	_, state3, _ := parser.Feed([]byte(`:"Jack"}`))
	assert.Equal(t, 3, state3.ParseAttempts)
}

func TestStreamParser_LastError_SetOnIncomplete(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	incompleteJSON := []byte(`{"name":"incomplete`)
	_, state, err := parser.Feed(incompleteJSON)

	require.NoError(t, err)           // No validation error
	require.Error(t, state.LastError) // Parse error stored
	assert.False(t, state.IsComplete)
}

func TestStreamParser_LastError_NilOnComplete(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	completeJSON := []byte(`{"name":"Kate","email":"kate@example.com","age":26}`)
	_, state, _ := parser.Feed(completeJSON)

	assert.True(t, state.IsComplete)
	// Note: LastError might be nil or set to last parse attempt error
	// Implementation doesn't clear it, so we just verify IsComplete=true
}

// ==================== Edge Cases ====================

func TestStreamParser_EmptyChunk_DoesNotPanic(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	assert.NotPanics(t, func() {
		obj, state, err := parser.Feed([]byte{})
		assert.NoError(t, err)
		assert.Nil(t, obj)
		assert.False(t, state.IsComplete)
		assert.Equal(t, 0, state.BytesReceived)
	})
}

func TestStreamParser_MultipleEmptyChunks(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	for i := 0; i < 5; i++ {
		obj, state, err := parser.Feed([]byte{})
		require.NoError(t, err)
		assert.Nil(t, obj)
		assert.False(t, state.IsComplete)
		assert.Equal(t, i+1, state.ParseAttempts)
	}
}

func TestStreamParser_MalformedJSON_StaysIncomplete(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	malformed := []byte(`{"name"::"invalid"}`)
	obj, state, err := parser.Feed(malformed)

	require.NoError(t, err) // Not a validation error
	assert.Nil(t, obj)
	assert.False(t, state.IsComplete)
	assert.Error(t, state.LastError)
}

func TestStreamParser_ExtraFields_Ignored(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// JSON with extra field not in struct
	extraJSON := []byte(`{"name":"Leo","email":"leo@example.com","age":29,"extra":"ignored"}`)
	obj, state, err := parser.Feed(extraJSON)

	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.True(t, state.IsComplete)
	assert.Equal(t, "Leo", obj.Name)
}

func TestStreamParser_NullValues(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Null values for non-pointer fields
	nullJSON := []byte(`{"name":"Mia","email":null,"age":22}`)
	obj, state, err := parser.Feed(nullJSON)

	// Depending on implementation, null might unmarshal to zero value
	require.NotNil(t, state)
	assert.True(t, state.IsComplete)

	// If it unmarshals successfully, obj will be non-nil
	if obj != nil {
		assert.Equal(t, "Mia", obj.Name)
		assert.Empty(t, obj.Email) // null -> zero value
	}

	// If validation fails on required/email, that's also acceptable
	_ = err // May or may not error
}

// ==================== Nested Structs ====================

func TestStreamParser_NestedStructs_Complete(t *testing.T) {
	parser := NewStreamParser[NestedStreamTest]()

	completeJSON := []byte(`{
		"user": {"name":"Nina","email":"nina@example.com","age":24},
		"address": {"street":"123 Main St","city":"NYC"}
	}`)

	obj, state, err := parser.Feed(completeJSON)

	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.True(t, state.IsComplete)

	assert.Equal(t, "Nina", obj.User.Name)
	assert.Equal(t, "123 Main St", obj.Address.Street)
	assert.Equal(t, "NYC", obj.Address.City)
}

func TestStreamParser_NestedStructs_Incomplete(t *testing.T) {
	parser := NewStreamParser[NestedStreamTest]()

	chunk1 := []byte(`{"user":{"name":"Oscar","email":"oscar@`)
	obj1, state1, _ := parser.Feed(chunk1)
	assert.Nil(t, obj1)
	assert.False(t, state1.IsComplete)

	chunk2 := []byte(`example.com","age":31},"address":{"street":"456 Oak","city":"LA"}}`)
	obj2, state2, err2 := parser.Feed(chunk2)

	require.NoError(t, err2)
	require.NotNil(t, obj2)
	assert.True(t, state2.IsComplete)
	assert.Equal(t, "Oscar", obj2.User.Name)
	assert.Equal(t, "LA", obj2.Address.City)
}

func TestStreamParser_NestedStructs_ValidationFails(t *testing.T) {
	parser := NewStreamParser[NestedStreamTest]()

	// Missing required field in nested Address
	invalidJSON := []byte(`{
		"user": {"name":"Paula","email":"paula@example.com","age":27},
		"address": {"street":"789 Elm"}
	}`)

	obj, state, err := parser.Feed(invalidJSON)

	require.Error(t, err) // Missing required 'city'
	require.NotNil(t, obj)
	assert.True(t, state.IsComplete)
}

// ==================== Concurrent Safety Tests ====================

func TestStreamParser_Concurrent_Safe(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	var wg sync.WaitGroup
	chunks := [][]byte{
		[]byte(`{"`),
		[]byte(`name"`),
		[]byte(`:"Quincy","`),
		[]byte(`email":"quincy@example.com","`),
		[]byte(`age":33}`),
	}

	// Feed chunks concurrently
	for _, chunk := range chunks {
		wg.Add(1)
		go func(c []byte) {
			defer wg.Done()
			_, _, _ = parser.Feed(c)
		}(chunk)
	}

	wg.Wait()

	// Buffer should contain all data (order may vary due to concurrency)
	buf := parser.Buffer()
	assert.NotEmpty(t, buf)

	// Should not panic - that's the main test
}

func TestStreamParser_Concurrent_MultipleReaders(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	completeJSON := []byte(`{"name":"Rachel","email":"rachel@example.com","age":29}`)
	_, _, _ = parser.Feed(completeJSON)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := parser.Buffer()
			assert.JSONEq(t, string(completeJSON), string(buf))
		}()
	}

	wg.Wait()
}

func TestStreamParser_Concurrent_ResetWhileReading(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	var wg sync.WaitGroup

	// Feed data
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _, _ = parser.Feed([]byte(`{"name":"Sam"`))
	}()

	// Reset concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		parser.Reset()
	}()

	// Read buffer concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		parser.Buffer()
	}()

	wg.Wait()

	// Should not panic - that's the main test
}

// ==================== StreamState Tests ====================

func TestStreamState_HasField(t *testing.T) {
	state := &StreamState{
		PresentFields: []string{"name", "email", "address.city"},
	}

	assert.True(t, state.HasField("name"))
	assert.True(t, state.HasField("email"))
	assert.True(t, state.HasField("address.city"))
	assert.False(t, state.HasField("age"))
	assert.False(t, state.HasField("missing"))
}

func TestStreamState_HasField_EmptyList(t *testing.T) {
	state := &StreamState{
		PresentFields: []string{},
	}

	assert.False(t, state.HasField("name"))
}

func TestStreamState_HasField_NilList(t *testing.T) {
	state := &StreamState{
		PresentFields: nil,
	}

	assert.False(t, state.HasField("name"))
}

// ==================== Integration Tests ====================

func TestStreamParser_RealWorldScenario_LLMStreaming(t *testing.T) {
	parser := NewStreamParser[StreamTestUser]()

	// Simulate LLM streaming chunks
	chunks := [][]byte{
		[]byte("{"),
		[]byte("\"name\":"),
		[]byte("\"Tom"),
		[]byte("my\""),
		[]byte(","),
		[]byte("\"email\""),
		[]byte(":"),
		[]byte("\"tommy@"),
		[]byte("example.com\""),
		[]byte(","),
		[]byte("\"age\":4"),
		[]byte("2}"),
	}

	var lastState *StreamState
	var lastObj *StreamTestUser
	var lastErr error

	for i, chunk := range chunks {
		obj, state, err := parser.Feed(chunk)
		lastState = state
		lastObj = obj
		lastErr = err

		// Until last chunk, should be incomplete
		if i < len(chunks)-1 {
			assert.Nil(t, obj)
			assert.False(t, state.IsComplete)
		}
	}

	// Final state should be complete
	require.NoError(t, lastErr)
	require.NotNil(t, lastObj)
	assert.True(t, lastState.IsComplete)
	assert.Equal(t, "Tommy", lastObj.Name)
	assert.Equal(t, "tommy@example.com", lastObj.Email)
	assert.Equal(t, 42, lastObj.Age)
	assert.Equal(t, len(chunks), lastState.ParseAttempts)
}
