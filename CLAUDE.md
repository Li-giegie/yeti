# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
go build ./...          # Build all packages
go test ./...           # Run all tests
go test -v ./...        # Run tests with verbose output
go test -run TestClient # Run a single test
```

## Architecture

This is an HTTP client library using a builder pattern. The core flow:

1. **Client** - Main entry point. Configure with `NewClient()`. Supports middleware-like hooks:
   - `AddBeforeRequest()` - Modify Requester before request is built
   - `AddAfterRequest()` - Modify http.Request after building but before sending
   - `AddAfterResponse()` - Process response before returning

2. **Requester** - Obtained via `client.R()`. Fluent builder for constructing requests:
   - Method: `SetMethod()`
   - URL: `SetUrl()`
   - Paths: `AddPath()` (appends path segments)
   - Query params: `AddQuery()`, `AddQueryAny()`
   - Headers: `SetHeader()`, `AddHeader()`, `SetHeaderAny()`, `AddHeaderAny()`
   - Body types: `SetBodyJSON()`, `SetBodyXML()`, `SetBodyForm()`, `SetBodyMultipartForm()`, `SetBodyText()`, `SetBodyBinary()`

3. **Response** - Returned by `DoResponse()`. Provides:
   - `StatusCodeEq()` - Assert status code
   - `JSON()` / `XML()` - Parse body into struct

## Debugging

Pass debug context values to enable request/response logging:
```go
ctx := context.WithValue(context.TODO(), client.ANNE_DEBUG, true)
// or
ctx := context.WithValue(context.TODO(), client.ANNE_REQUEST_DEBUG, true)
ctx := context.WithValue(context.TODO(), client.ANNE_RESPONSE_DEBUG, true)
// or use DoDebug() for shorthand
```

## Key Implementation Details

- `HttpClient` interface allows custom HTTP client injection
- `toString()` helper converts various types (primitives, pointers, reflect-based) to string for query/header values
- Path joining handles trailing/leading slash edge cases
- Multipart form files are read asynchronously via goroutine/pipe
