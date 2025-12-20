---
sidebar_position: 2
---

# LLM Streaming Integration

Validate structured JSON output from LLMs as it streams token-by-token.

## The Challenge

Large Language Models return streaming responses where JSON arrives piece by piece. This presents several challenges:

1. **Incomplete JSON**: You receive fragments like `{"name": "Al` before the full JSON is available
2. **Validation Timing**: When should you validate - partially or only at completion?
3. **Error Recovery**: What happens when the LLM returns invalid JSON?
4. **Schema Communication**: How do you tell the LLM what format you expect?

Pedantigo's `StreamParser` solves this by accumulating chunks until complete valid JSON arrives, then validating it fully.

## OpenAI Structured Output

OpenAI's function calling API returns streaming responses with structured data. Use Pedantigo to validate tool calls as they arrive:

```go
package main

import (
    "fmt"
    "github.com/smrutai/pedantigo"
)

type ToolCall struct {
    Name   string         `json:"name" pedantigo:"required"`
    Args   map[string]any `json:"args" pedantigo:"required"`
}

type FunctionResponse struct {
    ToolCalls []ToolCall `json:"tool_calls" pedantigo:"min_items=1,max_items=5"`
}

func main() {
    parser := pedantigo.NewStreamParser[FunctionResponse]()

    // Simulate streaming chunks from OpenAI
    chunks := []string{
        `{"tool_calls":[{"name":"search"`,
        `,"args":{"query":"Go validation"}`,
        `}]}`,
    }

    for _, chunk := range chunks {
        result, state, err := parser.Feed([]byte(chunk))

        if !state.IsComplete {
            fmt.Printf("Received %d bytes, waiting for more...\n", state.BytesReceived)
            continue
        }

        if err != nil {
            fmt.Printf("Validation failed: %v\n", err)
            continue
        }

        // Process validated tool calls
        for _, call := range result.ToolCalls {
            fmt.Printf("Execute: %s(%v)\n", call.Name, call.Args)
        }
    }
}
```

## Anthropic Claude Streaming

Anthropic's API streams text and structured outputs. Validate multi-field responses from Claude agents:

```go
type AgentResponse struct {
    Thought string `json:"thought" pedantigo:"required,min=10"`
    Action  string `json:"action" pedantigo:"required,oneof=search calculate respond think"`
    Query   string `json:"query" pedantigo:"min=1,max=500"`
    Result  string `json:"result"`
}

func processClaudeStream(chunks <-chan string) {
    parser := pedantigo.NewStreamParser[AgentResponse]()

    for chunk := range chunks {
        result, state, err := parser.Feed([]byte(chunk))

        if !state.IsComplete {
            continue
        }

        if err != nil {
            fmt.Println("Invalid response from Claude:", err)
            // Implement retry logic or ask Claude to reformat
            continue
        }

        // Use validated response
        switch result.Action {
        case "search":
            performSearch(result.Query)
        case "calculate":
            performCalculation(result.Query)
        case "respond":
            fmt.Println(result.Result)
        }

        parser.Reset() // Ready for next response
    }
}
```

## Providing Schema to LLMs

Include the JSON schema in your LLM prompt so it knows the expected format:

```go
type ExtractedData struct {
    Title       string   `json:"title" pedantigo:"required,max=200"`
    Description string   `json:"description" pedantigo:"max=2000"`
    Tags        []string `json:"tags" pedantigo:"max_items=10"`
    Confidence  float64  `json:"confidence" pedantigo:"ge=0,le=1"`
}

func getSystemPrompt() string {
    // Generate schema once (cached for performance)
    schema := pedantigo.Schema[ExtractedData]()

    schemaJSON, _ := json.MarshalIndent(schema, "", "  ")

    return fmt.Sprintf(`You are a data extraction assistant.
Extract the following fields from the input text.
Return valid JSON matching this schema:

%s

Requirements:
- Return only valid JSON, no markdown or explanations
- All required fields must be present
- Confidence should be between 0 and 1
`, string(schemaJSON))
}
```

## Handling Invalid Responses

LLMs occasionally generate invalid JSON. Implement graceful error handling:

```go
type ExtractionResult struct {
    Content string `json:"content" pedantigo:"required"`
    Score   float64 `json:"score" pedantigo:"ge=0,le=1"`
}

func extractWithRetry(llm *LLMClient, text string, maxRetries int) (*ExtractionResult, error) {
    parser := pedantigo.NewStreamParser[ExtractionResult]()

    for attempt := 0; attempt < maxRetries; attempt++ {
        parser.Reset()

        chunks := llm.Stream(text)
        for chunk := range chunks {
            result, state, err := parser.Feed([]byte(chunk))

            if !state.IsComplete {
                continue
            }

            if err == nil {
                return result, nil // Success
            }

            // Validation failed
            if attempt < maxRetries-1 {
                fmt.Printf("Attempt %d failed, retrying: %v\n", attempt+1, err)
                break // Try again
            } else {
                return nil, fmt.Errorf("max retries exceeded: %w", err)
            }
        }
    }

    return nil, errors.New("failed to extract data")
}
```

## Union Types for Multi-Action LLMs

For agents that return different response types based on an action field, use discriminated unions:

```go
type SearchAction struct {
    Action string   `json:"action" pedantigo:"required,const=search"`
    Query  string   `json:"query" pedantigo:"required,min=1"`
    Limit  int      `json:"limit" pedantigo:"min=1,max=100"`
}

type CalculateAction struct {
    Action     string `json:"action" pedantigo:"required,const=calculate"`
    Expression string `json:"expression" pedantigo:"required"`
}

type RespondAction struct {
    Action   string `json:"action" pedantigo:"required,const=respond"`
    Response string `json:"response" pedantigo:"required,min=1"`
}

func processAgentAction(jsonData []byte) {
    // The discriminator field "action" determines the type
    var action map[string]any
    json.Unmarshal(jsonData, &action)

    switch action["action"] {
    case "search":
        var searchReq SearchAction
        if result, errs := pedantigo.Unmarshal[SearchAction](jsonData); errs == nil {
            performSearch(result.Query, result.Limit)
        }

    case "calculate":
        var calcReq CalculateAction
        if result, errs := pedantigo.Unmarshal[CalculateAction](jsonData); errs == nil {
            computeExpression(result.Expression)
        }

    case "respond":
        var respReq RespondAction
        if result, errs := pedantigo.Unmarshal[RespondAction](jsonData); errs == nil {
            sendResponse(result.Response)
        }
    }
}
```

## Stream State Tracking

Monitor streaming progress with the `StreamState` object:

```go
type StreamState struct {
    // IsComplete indicates valid JSON has been received
    IsComplete bool

    // BytesReceived tracks total accumulated bytes
    BytesReceived int

    // ParseAttempts counts how many parse attempts were made
    ParseAttempts int

    // LastError stores the most recent parsing error
    LastError error
}

func monitorStream(parser *pedantigo.StreamParser[Response]) {
    for chunk := range getStreamChunks() {
        result, state, err := parser.Feed([]byte(chunk))

        // Log streaming metrics
        if state.ParseAttempts > 1 {
            fmt.Printf("Bytes: %d, Attempts: %d\n",
                state.BytesReceived, state.ParseAttempts)
        }

        if !state.IsComplete {
            continue
        }

        if err != nil {
            fmt.Printf("Validation error on attempt %d: %v\n",
                state.ParseAttempts, state.LastError)
            continue
        }

        // Process validated result
        processResponse(result)
    }
}
```

## Complete Agent Loop Example

A full example showing an agent that streams responses and validates them:

```go
type AgentStep struct {
    Reasoning string `json:"reasoning" pedantigo:"required,min=5"`
    Action    string `json:"action" pedantigo:"required,oneof=think search call respond complete"`
    Details   string `json:"details" pedantigo:"max=1000"`
}

type AgentLoop struct {
    llm    *LLMClient
    parser *pedantigo.StreamParser[AgentStep]
    maxSteps int
}

func NewAgentLoop(llm *LLMClient) *AgentLoop {
    return &AgentLoop{
        llm:      llm,
        parser:   pedantigo.NewStreamParser[AgentStep](),
        maxSteps: 10,
    }
}

func (al *AgentLoop) Run(input string) error {
    for step := 0; step < al.maxSteps; step++ {
        al.parser.Reset()

        response := al.llm.StreamResponse(input)

        var lastValid *AgentStep
        for chunk := range response {
            result, state, err := al.parser.Feed([]byte(chunk))

            if !state.IsComplete {
                continue
            }

            if err != nil {
                fmt.Printf("Step %d: Invalid response: %v\n", step, err)
                break
            }

            lastValid = result

            // Process the validated step
            switch result.Action {
            case "think":
                fmt.Printf("[THINK] %s\n", result.Reasoning)
                input = result.Details

            case "search":
                fmt.Printf("[SEARCH] %s\n", result.Details)
                results := performSearch(result.Details)
                input = fmt.Sprintf("Search results: %v", results)

            case "respond":
                fmt.Printf("[RESPONSE] %s\n", result.Details)
                return nil

            case "complete":
                fmt.Printf("[COMPLETE] %s\n", result.Reasoning)
                return nil
            }
        }

        if lastValid == nil {
            return errors.New("agent returned invalid response")
        }
    }

    return errors.New("max steps exceeded")
}
```

## Capturing Unexpected LLM Fields

LLMs may include fields you didn't request. Use [ExtraAllow](/docs/api/initialization#extra-allow) to capture these fields for debugging prompts and evaluating model accuracy:

```go
type LLMResponse struct {
    Answer     string         `json:"answer" pedantigo:"required"`
    Confidence float64        `json:"confidence"`
    Extras     map[string]any `json:"-" pedantigo:"extra_fields"` // Captures unexpected fields
}

func processLLMOutput(output []byte) {
    validator := pedantigo.New[LLMResponse](pedantigo.ValidatorOptions{
        ExtraFields: pedantigo.ExtraAllow,  // Capture, don't reject
    })

    response, err := validator.Unmarshal(output)
    if err != nil {
        log.Printf("Validation failed: %v", err)
        return
    }

    // Log extra fields for prompt improvement
    if len(response.Extras) > 0 {
        log.Printf("LLM included unexpected fields: %v", response.Extras)
        // Track metrics to evaluate model accuracy
        metrics.RecordExtraFields(modelName, promptID, response.Extras)
        // Example captured: {"reasoning": "...", "sources": [...]}
    }

    // Use validated response
    fmt.Printf("Answer: %s (%.2f confidence)\n", response.Answer, response.Confidence)
}
```

**Use cases:**
- **Prompt debugging**: Identify when models add unrequested fields
- **Model accuracy evaluation**: Track field adherence across prompts
- **Prompt iteration**: Use captured extras to refine instructions

See [ExtraAllow Use Cases](/docs/api/initialization#extra-allow-use-cases) for more patterns.

## Performance Tips

1. **Reuse Parsers**: Create the parser once and call `Reset()` between streams
2. **Schema Caching**: `pedantigo.Schema[T]()` is cached after first call (240x faster)
3. **Error Context**: `StreamState` provides bytes received and parse attempts for debugging
4. **Non-Blocking**: `Feed()` returns immediately with incomplete state, don't wait for completion

## When to Use Streaming vs Simple API

Use `StreamParser` when:
- Receiving JSON from LLMs or external APIs token-by-token
- Need to validate as data arrives
- Want to detect malformed JSON early
- Processing large responses where latency matters

Use Simple API (`pedantigo.Unmarshal`) when:
- You already have complete JSON
- Parsing from request bodies or files
- Simplicity is preferred over streaming semantics

:::tip
Combine streaming validation with schema generation for a complete LLM integration solution: provide the schema in prompts, validate responses as they stream.
:::
