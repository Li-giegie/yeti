package client

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.Client != nil {
		t.Fatal("default Client should be nil (uses http.DefaultClient)")
	}
}

func TestClientHooks(t *testing.T) {
	var beforeCalled, afterReqCalled, afterRespCalled bool

	client := NewClient()
	client.AddBeforeRequest(func(r *Requester) error {
		beforeCalled = true
		r.SetHeader("X-Before", "yes")
		return nil
	})
	client.AddAfterRequest(func(r *http.Request) error {
		afterReqCalled = true
		return nil
	})
	client.AddAfterResponse(func(r *http.Response) error {
		afterRespCalled = true
		return nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	_, err := client.R().
		SetMethod("GET").
		SetUrl(svr.URL + "/test").
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !beforeCalled {
		t.Error("beforeRequest hook was not called")
	}
	if !afterReqCalled {
		t.Error("afterRequest hook was not called")
	}
	if !afterRespCalled {
		t.Error("afterResponse hook was not called")
	}
}

func TestRequesterSetMethod(t *testing.T) {
	client := NewClient()
	r := client.R().SetMethod("POST")
	if r.Method != "POST" {
		t.Errorf("expected POST, got %s", r.Method)
	}
}

func TestRequesterSetUrl(t *testing.T) {
	client := NewClient()
	r := client.R().SetUrl("http://example.com/api")
	if r.URL != "http://example.com/api" {
		t.Errorf("expected http://example.com/api, got %s", r.URL)
	}
}

func TestRequesterAddPath(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{"single path", []string{"/users"}, "/users"},
		{"multiple paths", []string{"/users", "/123"}, "/users/123"},
		{"with leading slash", []string{"users"}, "/users"},
		{"empty path ignored", []string{"/users", ""}, "/users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			req, err := client.R().SetMethod("GET").SetUrl(svr.URL).Do(context.TODO())
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			req.Body.Close()
			_ = tt // use test case
		})
	}
}

func TestRequesterAddQuery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") != "test" {
			t.Errorf("expected name=test, got %v", r.URL.Query())
		}
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("GET").
		SetUrl(svr.URL+"/test").
		AddQuery("name", "test").
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestRequesterSetHeader(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Errorf("expected X-Custom=value, got %v", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("GET").
		SetUrl(svr.URL+"/test").
		SetHeader("X-Custom", "value").
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestRequesterSetBodyJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("POST").
		SetUrl(svr.URL + "/test").
		SetBodyJSON(map[string]string{"key": "value"}).
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestRequesterSetBodyXML(t *testing.T) {
	type xmlData struct {
		Name string `xml:"name"`
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/xml" {
			t.Errorf("expected application/xml, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("POST").
		SetUrl(svr.URL + "/test").
		SetBodyXML(xmlData{Name: "test"}).
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestRequesterSetBodyForm(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("POST").
		SetUrl(svr.URL + "/test").
		SetBodyForm(url.Values{"name": {"value"}}).
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestRequesterSetBodyText(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("expected text/plain, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("POST").
		SetUrl(svr.URL + "/test").
		SetBodyText("hello world").
		Do(context.TODO())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestResponseStatusCodeEq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/created", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	tests := []struct {
		name    string
		path    string
		codes   int
		wantErr bool
	}{
		{"status 200 matches", "/ok", 200, false},
		{"status 201 matches", "/created", 201, false},
		{"no match", "/ok", 404, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			resp := client.R().
				SetMethod("GET").
				SetUrl(svr.URL + tt.path).
				DoResponse(context.TODO())
			resp.EqStatusCode(tt.codes)
			if (resp.Err != nil) != tt.wantErr {
				t.Errorf("StatusCodeEq() error = %v, wantErr %v", resp.Err, tt.wantErr)
			}
		})
	}
}

func TestResponseJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name":"test","value":123}`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp := client.R().
		SetMethod("GET").
		SetUrl(svr.URL + "/test").
		DoResponse(context.TODO())

	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	err := resp.JSON(&result)
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}
	if result.Name != "test" || result.Value != 123 {
		t.Errorf("expected {test, 123}, got {%s, %d}", result.Name, result.Value)
	}
}

func TestResponseXML(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<item><name>test</name><value>123</value></item>`))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp := client.R().
		SetMethod("GET").
		SetUrl(svr.URL + "/test").
		DoResponse(context.TODO())

	var result struct {
		XMLName xml.Name `xml:"item"`
		Name    string   `xml:"name"`
		Value   int      `xml:"value"`
	}
	err := resp.XML(&result)
	if err != nil {
		t.Fatalf("XML() failed: %v", err)
	}
	if result.Name != "test" || result.Value != 123 {
		t.Errorf("expected {test, 123}, got {%s, %d}", result.Name, result.Value)
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantOk  bool
		wantVal string
	}{
		{"string", "hello", true, "hello"},
		{"int", 42, true, "42"},
		{"int64", int64(1234567890), true, "1234567890"},
		{"float64", 3.14, true, "3.14"},
		{"bool true", true, true, "true"},
		{"bool false", false, true, "false"},
		{"[]byte", []byte("test"), true, "test"},
		{"nil", nil, false, ""},
		{"pointer to string", func() any { s := "ptr"; return &s }(), true, "ptr"},
		{"pointer to int", func() any { i := 99; return &i }(), true, "99"},
		{"unsupported type", struct{}{}, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toString(tt.input)
			if ok != tt.wantOk {
				t.Errorf("toString() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.wantVal {
				t.Errorf("toString() = %v, want %v", got, tt.wantVal)
			}
		})
	}
}

func TestRequesterAddQueryAny(t *testing.T) {
	client := NewClient()
	r := client.R().AddQueryAny("key", 123)
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}
	if r.Query.Get("key") != "123" {
		t.Errorf("expected 123, got %s", r.Query.Get("key"))
	}
}

func TestRequesterAddHeaderAny(t *testing.T) {
	client := NewClient()
	r := client.R().AddHeaderAny("X-Value", 42)
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}
	if r.Header.Get("X-Value") != "42" {
		t.Errorf("expected 42, got %s", r.Header.Get("X-Value"))
	}
}

func TestRequesterSetHeaderAny(t *testing.T) {
	client := NewClient()
	r := client.R().SetHeaderAny("X-Value", 42)
	if r.Err != nil {
		t.Errorf("unexpected error: %v", r.Err)
	}
	if r.Header.Get("X-Value") != "42" {
		t.Errorf("expected 42, got %s", r.Header.Get("X-Value"))
	}
}

func TestRequesterDoResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "header-value")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("response body"))
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp := client.R().
		SetMethod("GET").
		SetUrl(svr.URL + "/test").
		DoResponse(context.TODO())

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Custom") != "header-value" {
		t.Errorf("expected X-Custom=header-value, got %s", resp.Header.Get("X-Custom"))
	}
	if string(resp.Body) != "response body" {
		t.Errorf("expected 'response body', got %s", string(resp.Body))
	}
}

func TestRequesterReset(t *testing.T) {
	client := NewClient()
	r := client.R().
		SetMethod("POST").
		SetUrl("http://example.com").
		SetHeader("X-Test", "value").
		SetBodyJSON(map[string]any{"key": "value"})

	r.Reset()

	if r.Err != nil {
		t.Errorf("Reset() Err = %v, want nil", r.Err)
	}
	if r.Method != "" {
		t.Errorf("Reset() Method = %s, want empty", r.Method)
	}
	if r.URL != "" {
		t.Errorf("Reset() URL = %s, want empty", r.URL)
	}
}

func TestErrorShortCircuit(t *testing.T) {
	client := NewClient()
	client.AddBeforeRequest(func(r *Requester) error {
		return errors.New("before request error")
	})

	resp, err := client.R().
		SetMethod("GET").
		SetUrl("http://example.com").
		Do(context.TODO())

	if err == nil {
		t.Error("expected error, got nil")
	}
	if resp != nil {
		t.Error("expected nil response on error")
	}
}

func TestDoDebug(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	client := NewClient()
	resp, err := client.R().
		SetMethod("GET").
		SetUrl(svr.URL + "/test").
		DoDebug(context.TODO())
	if err != nil {
		t.Fatalf("DoDebug() failed: %v", err)
	}
	resp.Body.Close()
}

func TestGetBody(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	fmt.Println(readBody(req))

	req, _ = http.NewRequest("POST", "http://example.com", bytes.NewBufferString("NewBufferString"))
	fmt.Println(readBody(req))

	req, _ = http.NewRequest("POST", "http://example.com", bytes.NewBuffer([]byte("NewBuffer")))
	fmt.Println(readBody(req))

	req, _ = http.NewRequest("POST", "http://example.com", bytes.NewReader([]byte("NewReader")))
	fmt.Println(readBody(req))

	file, _ := os.Open("./README_CN.md")
	defer file.Close()
	req, _ = http.NewRequest("POST", "http://example.com", file)
	fmt.Println(readBody(req))

}

func readBody(req *http.Request) string {
	if req.GetBody == nil {
		return "<nil>"
	}
	r, err := req.GetBody()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func newBodyBuffer(r io.Reader, size int) *bodyBuffer {
	return &bodyBuffer{
		r:    r,
		size: size,
	}
}

type bodyBuffer struct {
	r      io.Reader
	size   int
	buffer []byte
	file   *os.File
}

func (b *bodyBuffer) Read(p []byte) (n int, err error) {
	var fErr error
	n, err = b.r.Read(p)
	if b.file != nil { // 写入到文件
		if n, fErr = b.file.Write(p[:n]); fErr != nil {
			return
		}
	} else { // 写入到内存
		if b.size > 0 && len(b.buffer)+len(p) > b.size {
			if b.file, fErr = os.CreateTemp("", ".*.body"); fErr != nil {
				return
			}
			if n, fErr = b.file.Write(b.buffer); fErr != nil {
				return
			}
			if n, fErr = b.file.Write(p[:n]); fErr != nil {
				return
			}
			b.buffer = nil
		} else {
			b.buffer = append(b.buffer, p[:n]...)
		}
	}
	return
}

func (b *bodyBuffer) Close() error {
	if b.file != nil {
		b.file.Close()
		os.Remove(b.file.Name())
	}
	return nil
}

func TestBodyBuffer(t *testing.T) {
	rc := newBodyBuffer(bytes.NewBufferString("NewBufferString"), 4)
	b := make([]byte, 4)
	for {
		n, err := rc.Read(b)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(string(b[:n]))
	}
}
