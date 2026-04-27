# Yeti HTTP Client

[中文文档](./README_CN.md)

A Go HTTP client library with a fluent Builder pattern.

## Install

```bash
go get github.com/Li-giegie/yeti
```

## Quick Start

```go
import "github.com/Li-giegie/yeti"

// Create client
c := yeti.NewClient()

// Send GET request
var result MyStruct
c.Get().
    SetUrl("https://api.example.com/users").
    AddPath("123").
    AddQuery("fields", "name,email").
    DoResponse(ctx).
    JSON(&result)
```

## Core Components

### Client

Main entry point, created via `NewClient()`.

```go
c := yeti.NewClient()
```

Supports hook methods to intercept requests/responses:

- `AddBeforeRequest(fn)` - Modify Requester before building request
- `AddAfterRequest(fn)` - Modify `*http.Request` before sending
- `AddAfterResponse(fn)` - Process response before returning

### Requester

Obtained via `yeti.R()` or shortcut methods (`Get()`, `Post()`, `Put()`, `Patch()`, `Delete()`, etc.).

Builder methods:

| Method | Description |
|--------|-------------|
| `SetMethod(method)` | Set HTTP method |
| `SetUrl(url)` | Set request URL |
| `AddPath(segment)` | Add URL path segment (auto handles slashes) |
| `AddQuery(key, value)` | Set query parameter |
| `AddQueryAny(key, value)` | Add query parameter with auto type conversion |
| `SetHeader(key, value)` | Set header (overwrite) |
| `AddHeader(key, value)` | Add header (appendable) |
| `SetHeaderAny(key, value)` | Set header with auto type conversion |
| `AddHeaderAny(key, value)` | Add header with auto type conversion |

Request body methods:

| Method | Content-Type |
|--------|--------------|
| `SetBodyJSON(v)` | application/json |
| `SetBodyXML(v)` | application/xml |
| `SetBodyForm(url.Values)` | application/x-www-form-urlencoded |
| `SetBodyMultipartForm(map)` | multipart/form-data |
| `SetBodyText(string)` | text/plain |
| `SetBodyBinary(io.Reader)` | application/octet-stream |

### Response

Returned by `DoResponse()`, provides response handling:

| Method | Description |
|--------|-------------|
| `EqStatusCode(code)` | Assert status code |
| `JSON(v)` | Parse JSON into struct |
| `XML(v)` | Parse XML into struct |
| `Validate(fn)` | Custom validation |

## Debugging

Enable debug logging via context:

```go
ctx := context.WithValue(context.TODO(), yeti.ANNE_DEBUG, true)
// Or enable separately
ctx := context.WithValue(context.TODO(), yeti.ANNE_REQUEST_DEBUG, true)
ctx := context.WithValue(context.TODO(), yeti.ANNE_RESPONSE_DEBUG, true)

// Use DoDebug() shorthand
resp, err := requester.DoDebug(ctx)
```

## Examples

### Basic Requests

```go
// GET request
c.Get().
    SetUrl("https://api.example.com").
    AddPath("users").
    AddQuery("page", 1).
    DoResponse(ctx)

// POST JSON
c.Post().
    SetUrl("https://api.example.com/users").
    SetBodyJSON(User{Name: "Alice", Email: "alice@example.com"}).
    DoResponse(ctx)

// File upload
c.Post().
    SetUrl("https://api.example.com/upload").
    SetBodyMultipartForm(map[string]any{
        "name": "my-file",
        "file": os.Open("test.png"),
    }).
    DoResponse(ctx)
```

### Custom HTTP Client

```go
c := yeti.NewClient()
c.Client = &http.Client{Timeout: 10 * time.Second}
```

### Using Hooks

```go
c := yeti.NewClient()

// Add auth header before request
c.AddBeforeRequest(func(r *yeti.Requester) error {
    r.SetHeader("Authorization", "Bearer token")
    return nil
})

// Log response after response
c.AddAfterResponse(func(resp *http.Response) error {
    log.Printf("Response: %s", resp.Status)
    return nil
})
```

## Build Commands

```bash
go build ./...     # Build all packages
go test ./...      # Run all tests
go test -v ./...   # Verbose output
```
