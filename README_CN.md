# Yeti HTTP Client

Go 语言 HTTP 客户端库，采用流畅的构建器（Builder）模式。

## 安装

```bash
go get github.com/Li-egiegie/yeti
```

## 快速开始

```go
import "github.com/your/module/yeti/client"

// 创建客户端
c := client.NewClient()

// 发送 GET 请求
var result MyStruct
c.Get().
    SetUrl("https://api.example.com/users").
    AddPath("123").
    AddQuery("fields", "name,email").
    DoResponse(ctx).
    JSON(&result)
```

## 核心组件

### Client

主入口点，通过 `NewClient()` 创建。

```go
c := client.NewClient(
    client.WithTimeout(30 * time.Second),
)
```

支持以下钩子方法拦截请求/响应：

- `AddBeforeRequest(fn)` - 构建请求前修改 Requester
- `AddAfterRequest(fn)` - 发送前修改 `*http.Request`
- `AddAfterResponse(fn)` - 处理响应后修改 `*http.Response`

### Requester

通过 `client.R()` 或快捷方法（`Get()`、`Post()`、`Put()`、`Patch()`、`Delete()` 等）获取。

支持的构建方法：

| 方法 | 说明 |
|------|------|
| `SetMethod(method)` | 设置 HTTP 方法 |
| `SetUrl(url)` | 设置请求 URL |
| `AddPath(segment)` | 添加 URL 路径段（自动处理斜杠） |
| `AddQuery(key, value)` | 设置查询参数 |
| `AddQueryAny(key, value)` | 自动类型转换后添加查询参数 |
| `SetHeader(key, value)` | 设置请求头（覆盖） |
| `AddHeader(key, value)` | 添加请求头（可重复） |
| `SetHeaderAny(key, value)` | 自动类型转换后设置请求头 |
| `AddHeaderAny(key, value)` | 自动类型转换后添加请求头 |

支持的请求体：

| 方法 | Content-Type |
|------|--------------|
| `SetBodyJSON(v)` | application/json |
| `SetBodyXML(v)` | application/xml |
| `SetBodyForm(url.Values)` | application/x-www-form-urlencoded |
| `SetBodyMultipartForm(map)` | multipart/form-data |
| `SetBodyText(string)` | text/plain |
| `SetBodyBinary(io.Reader)` | application/octet-stream |

### Response

通过 `DoResponse()` 返回，提供响应处理方法：

| 方法 | 说明 |
|------|------|
| `EqStatusCode(code)` | 断言状态码 |
| `JSON(v)` | 解析 JSON 到结构体 |
| `XML(v)` | 解析 XML 到结构体 |
| `Validate(fn)` | 自定义验证 |

## 调试

通过 context 启用调试日志：

```go
ctx := context.WithValue(context.TODO(), client.ANNE_DEBUG, true)
// 或分别启用
ctx := context.WithValue(context.TODO(), client.ANNE_REQUEST_DEBUG, true)
ctx := context.WithValue(context.TODO(), client.ANNE_RESPONSE_DEBUG, true)

// 使用 DoDebug() 快捷方法
resp, err := requester.DoDebug(ctx)
```

## 示例

### 基础请求

```go
// GET 请求
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

// 文件上传
c.Post().
    SetUrl("https://api.example.com/upload").
    SetBodyMultipartForm(map[string]any{
        "name": "my-file",
        "file": os.Open("test.png"),
    }).
    DoResponse(ctx)
```

### 自定义 HTTP 客户端

```go
c := client.NewClient()
c.Client = &http.Client{Timeout: 10 * time.Second}
```

### 使用钩子

```go
c := client.NewClient()

// 请求前添加认证头
c.AddBeforeRequest(func(r *client.Requester) error {
    r.SetHeader("Authorization", "Bearer token")
    return nil
})

// 响应后记录日志
c.AddAfterResponse(func(resp *http.Response) error {
    log.Printf("Response: %s", resp.Status)
    return nil
})
```

## 构建命令

```bash
go build ./...     # 构建所有包
go test ./...      # 运行所有测试
go test -v ./...   # 详细输出
```
