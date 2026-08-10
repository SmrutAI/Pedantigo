---
sidebar_position: 5
---

# Streaming Validation

Validate JSON as it streams in from LLMs and other real-time sources.

## The Problem

Large Language Models like OpenAI and Anthropic stream responses token-by-token. Your application needs to:

1. Accumulate tokens into valid JSON
2. Validate the structure once complete
3. Avoid parsing errors on incomplete JSON
4. Handle errors gracefully if the stream is interrupted

Pedantigo's `StreamParser` solves this by buffering chunks and validating only when the JSON is structurally complete.

:::tip
For complete LLM integration examples, see [LLM Streaming Examples](/docs/examples/llm-streaming). For generating schemas to send to LLMs, see [JSON Schema Generation](/docs/concepts/schema).
:::

## Creating a StreamParser

### Basic Usage

The simplest way to create a stream parser:

```go
type Message struct {
    Role    string `json:"role" validate:"required,oneof=user assistant system"`
    Content string `json:"content" validate:"required,min=1"`
}

// Create a parser for streaming messages
parser := validator.NewStreamParser[Message]()
```

### With Custom Validator Options

For advanced validation configurations:

```go
parser := validator.NewStreamParser[Message](validator.Options{
    // Custom options if needed
})
```

### With Custom Validator (for unions)

If you need discriminated union support or other advanced features, create a custom validator first:

```go
type MessageContent struct {
    Type string `json:"type" validate:"required,oneof=text image video"`
    Body string `json:"body" validate:"required"`
}

// Create a validator with custom configuration
messageContentValidator := validator.New[MessageContent]()

// Pass it to the stream parser
parser := validator.NewStreamParserWithValidator[MessageContent](messageContentValidator)
```

## Feeding Chunks

### The Feed Method

As chunks arrive from the stream, pass them to the parser:

```go
// Feed returns (result, state, error)
result, state, err := parser.Feed(chunk)

// While JSON is incomplete:
// - result is nil
// - err is nil (not an error to receive incomplete JSON)
// - state.IsComplete is false

// Once JSON is complete and valid:
// - result contains the validated struct
// - state.IsComplete is true
// - err is nil if validation passed
```

**Important**: Feed does NOT treat incomplete JSON as an error. It accumulates chunks until valid JSON is available.

### StreamState Explained

Each call to `Feed()` returns a `StreamState` object:

```go
type StreamState struct {
    BytesReceived int   // Total bytes accumulated so far
    ParseAttempts int   // Number of parse attempts made
    IsComplete    bool  // Whether we have valid, complete JSON
    LastError     error // Last parsing error (if incomplete)
}
```

Use `state.IsComplete` to detect when validation has completed:

```go
for chunk := range streamChan {
    result, state, err := parser.Feed(chunk)

    if state.IsComplete {
        // JSON is complete and validated
        if err != nil {
            // Validation failed
            fmt.Printf("Validation error: %v\n", err)
        } else {
            // result is valid and ready to use
            fmt.Printf("Valid message: %v\n", result)
        }
        break
    }

    // Not complete yet, keep feeding chunks
    fmt.Printf("Accumulated %d bytes...\n", state.BytesReceived)
}
```

## Handling Incomplete Streams

If the stream terminates before sending complete JSON, check the state:

```go
parser := validator.NewStreamParser[Message]()

// Stream closes unexpectedly
var lastState *validator.StreamState
for chunk := range streamChan {
    result, state, err := parser.Feed([]byte(chunk))
    lastState = state
    if state.IsComplete {
        // Got valid JSON before stream closed
        break
    }
}

// After stream closes, check if we have incomplete data
state := lastState
if !state.IsComplete && state.BytesReceived > 0 {
    fmt.Printf("Stream ended with incomplete JSON (%d bytes)\n", state.BytesReceived)
    fmt.Printf("Last parse error: %v\n", state.LastError)
}
```

## Complete LLM Example

### OpenAI Streaming Response

```go
package main

import (
    "context"
    "fmt"
    "io"

    openai "github.com/sashabaranov/go-openai"
    "github.com/SmrutAI/pedantigo/v2/validator"
)

type ChatCompletion struct {
    Content string `json:"content" validate:"required"`
    Model   string `json:"model" validate:"required"`
}

func handleOpenAIStream(ctx context.Context, client *openai.Client) error {
    parser := validator.NewStreamParser[ChatCompletion]()

    stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "What is 2+2?",
            },
        },
    })
    if err != nil {
        return err
    }
    defer stream.Close()

    var lastState *validator.StreamState
    for {
        response, err := stream.Recv()
        if err == io.EOF {
            // Stream ended
            if lastState != nil && !lastState.IsComplete {
                fmt.Printf("Warning: Stream ended with incomplete JSON\n")
            }
            break
        }
        if err != nil {
            return err
        }

        // The response chunks need to be assembled into JSON
        chunk := response.Choices[0].Delta.Content

        result, state, parseErr := parser.Feed([]byte(chunk))
        lastState = state

        if state.IsComplete {
            if parseErr != nil {
                fmt.Printf("Validation error: %v\n", parseErr)
                return parseErr
            }
            fmt.Printf("Valid completion: %v\n", result.Content)
            break
        }
    }

    return nil
}
```

### Anthropic Streaming Response

```go
package main

import (
    "context"
    "fmt"

    "github.com/anthropics/anthropic-sdk-go"
    "github.com/SmrutAI/pedantigo/v2/validator"
)

type ContentBlock struct {
    Type string `json:"type" validate:"required,oneof=text image"`
    Text string `json:"text" validate:"required,min=1"`
}

func handleAnthropicStream(ctx context.Context, client *anthropic.Client) error {
    parser := validator.NewStreamParser[ContentBlock]()

    stream := client.Messages.NewStream(ctx, anthropic.MessageNewParams{
        Model: anthropic.ModelClaude3Sonnet20240229,
        MaxTokens: anthropic.Int(1024),
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock("What is 2+2?")),
        },
    })
    defer stream.Close()

    for stream.Next() {
        event := stream.Current()

        // Extract text chunk from the streaming event
        if event.Type == "content_block_delta" {
            delta := event.Delta

            result, state, parseErr := parser.Feed([]byte(delta))

            if state.IsComplete {
                if parseErr != nil {
                    fmt.Printf("Validation error: %v\n", parseErr)
                    return parseErr
                }
                fmt.Printf("Valid response: %v\n", result.Text)
                return nil
            }
        }
    }

    return stream.Err()
}
```

## Resetting Between Streams

Once a stream completes, reset the parser for the next stream:

```go
parser := validator.NewStreamParser[Message]()

// First stream
for chunk := range stream1 {
    result, state, err := parser.Feed(chunk)
    if state.IsComplete {
        handleMessage(result, err)
        break
    }
}

// Reset for next stream
parser.Reset()

// Second stream
for chunk := range stream2 {
    result, state, err := parser.Feed(chunk)
    if state.IsComplete {
        handleMessage(result, err)
        break
    }
}
```

## Getting the Buffer (Debug)

For debugging or inspecting accumulated data:

```go
parser := validator.NewStreamParser[Message]()

// Feed some chunks
parser.Feed([]byte(`{"role":"assi`))
parser.Feed([]byte(`stant","content":"Hello`))

// Get current buffer
buffer := parser.Buffer()
fmt.Println(string(buffer)) // {"role":"assistant","content":"Hello
```

This is useful for logging or understanding why a stream failed to complete.

## Best Practices

### 1. Always Check IsComplete

Don't assume validation succeeded just because `Feed()` didn't error:

```go
// Wrong: assumes validation is complete after any error
result, _, err := parser.Feed(chunk)
if err != nil {
    // err might be a validation error OR a parse error
}

// Correct: check if JSON is actually complete
result, state, err := parser.Feed(chunk)
if state.IsComplete {
    // Now we know JSON is complete
    if err != nil {
        // This is a validation error
    }
}
```

### 2. Handle Partial Streams Gracefully

Not all streams will send complete JSON:

```go
result, state, err := parser.Feed(chunk)

if state.IsComplete {
    // Process result
    fmt.Printf("Message: %v\n", result)
} else if state.BytesReceived > 1024*1024 { // 1MB limit
    return fmt.Errorf("stream exceeded size limit")
} else {
    // Continue feeding
}
```

### 3. Validate Error Messages

JSON parse errors can be informative:

```go
result, state, err := parser.Feed(chunk)

if state.IsComplete && err != nil {
    if validationErr, ok := err.(*validator.ValidationError); ok {
        // Handle field-level validation errors
        for _, fieldErr := range validationErr.Errors {
            fmt.Printf("Field %s: %s\n", fieldErr.Field, fieldErr.Message)
        }
    } else {
        // Handle parse errors
        fmt.Printf("Parse error: %v\n", err)
    }
}
```

### 4. Reset Between Different Streams

Always reset to avoid cross-contamination:

```go
parser := validator.NewStreamParser[Message]()

// Process stream 1
for chunk := range stream1 {
    result, state, _ := parser.Feed(chunk)
    if state.IsComplete {
        break
    }
}

// Must reset before stream 2
parser.Reset()

// Process stream 2
for chunk := range stream2 {
    // ...
}
```

### 5. Set Size Limits for Production

Prevent memory exhaustion from unbounded streams:

```go
const maxStreamSize = 10 * 1024 * 1024 // 10MB

result, state, _ := parser.Feed(chunk)

if state.BytesReceived > maxStreamSize {
    return fmt.Errorf("stream size exceeded limit")
}
```

## Key Differences from Unmarshal

`StreamParser` differs from `Validator.Unmarshal()` in important ways:

| Aspect | Unmarshal | StreamParser |
|--------|-----------|--------------|
| Input | Complete JSON bytes | Chunks accumulated over time |
| Error handling | Returns immediately on error | Continues accumulating until complete |
| Default behavior | Applies defaults automatically | Applies defaults only after complete |
| Use case | One-shot parsing | Real-time streaming (LLMs, feeds) |
| Thread safety | Not thread-safe | Thread-safe with internal mutex |

Choose `Unmarshal` for complete JSON data and `StreamParser` for incremental streams.

## Thread Safety

`StreamParser` is thread-safe internally but maintains a single buffer per instance. Do not share a parser across multiple goroutines feeding different streams:

```go
// Wrong: two goroutines feeding to same parser
parser := validator.NewStreamParser[Message]()
go func() {
    parser.Feed(stream1Chunk) // Race condition!
}()
go func() {
    parser.Feed(stream2Chunk) // Race condition!
}()

// Correct: separate parsers per stream
parser1 := validator.NewStreamParser[Message]()
parser2 := validator.NewStreamParser[Message]()

go func() {
    parser1.Feed(stream1Chunk)
}()
go func() {
    parser2.Feed(stream2Chunk)
}()
```

## Common Pitfalls

:::warning Watch Out For These

1. **Trusting `result` before `IsComplete`** - Feed() returns partial results. Only use `result` when `state.IsComplete == true`.

2. **Assuming `err == nil` means complete** - No error just means no parse failure yet. Always check `state.IsComplete`.

3. **Forgetting to Reset()** - Parsers accumulate data. Call `parser.Reset()` before processing a new stream.

4. **Unbounded memory growth** - Check `state.BytesReceived` and set limits in production.

5. **Processing the same completion twice** - Once `IsComplete` is true, break the loop. Don't keep feeding chunks.

:::

:::tip
For streaming with discriminated unions (polymorphic responses), see [Discriminated Unions - Streaming](/docs/concepts/unions#streaming-discriminated-unions).
:::

## Next Steps

- Learn about [Validation Basics](/docs/concepts/validation) for field constraints
- See [Constraints Reference](/docs/constraints/string) for available validation rules
- Check [Examples](/docs/examples/llm-streaming) for real-world streaming patterns
