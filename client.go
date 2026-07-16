package yeti

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	ANNE_DEBUG          = "ANNE_DEBUG"
	ANNE_REQUEST_DEBUG  = "ANNE_REQUEST_DEBUG"
	ANNE_RESPONSE_DEBUG = "ANNE_RESPONSE_DEBUG"
	BIND_QUERY_TAG      = "query"
)

var (
	DebugCtx         = context.WithValue(context.TODO(), ANNE_DEBUG, true)
	DebugRequestCtx  = context.WithValue(context.TODO(), ANNE_REQUEST_DEBUG, true)
	DebugResponseCtx = context.WithValue(context.TODO(), ANNE_RESPONSE_DEBUG, true)
)

type Option func(client *Client)

func NewClient(opts ...Option) *Client {
	var client Client
	for _, fn := range opts {
		fn(&client)
	}
	return &client
}

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}
type Client struct {
	Client        HttpClient
	BeforeRequest []func(*Requester) error
	AfterRequest  []func(*http.Request) error
	AfterResponse []func(*http.Response) error
	QueryTag      string
}

func (c *Client) AddBeforeRequest(fn func(*Requester) error) {
	c.BeforeRequest = append(c.BeforeRequest, fn)
}

func (c *Client) AddAfterRequest(fn func(*http.Request) error) {
	c.AfterRequest = append(c.AfterRequest, fn)
}

func (c *Client) AddAfterResponse(fn func(*http.Response) error) {
	c.AfterResponse = append(c.AfterResponse, fn)
}

type RequesterOption func(r *Requester)

func (c *Client) R(opts ...RequesterOption) *Requester {
	req := &Requester{
		Header: make(http.Header),
		client: c,
		Query:  make(url.Values),
	}
	for _, fn := range opts {
		fn(req)
	}
	return req
}

func (c *Client) Get() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodGet
	})
}

func (c *Client) HEAD() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodHead
	})
}

func (c *Client) Post() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodPost
	})
}

func (c *Client) Put() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodPut
	})
}

func (c *Client) Patch() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodPatch
	})
}

func (c *Client) Delete() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodDelete
	})
}

func (c *Client) Connect() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodConnect
	})
}

func (c *Client) Options() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodOptions
	})
}

func (c *Client) Trace() *Requester {
	return c.R(func(r *Requester) {
		r.Method = http.MethodTrace
	})
}

type Requester struct {
	Err    error
	Method string
	URL    string
	Paths  []string
	Query  url.Values
	Header http.Header
	Body   io.Reader
	client *Client
}

func (b *Requester) SetMethod(method string) *Requester {
	if b.Err != nil {
		return b
	}
	b.Method = method
	return b
}

func (b *Requester) SetUrl(u string) *Requester {
	if b.Err != nil {
		return b
	}
	b.URL = u
	return b
}

func (b *Requester) AddPath(p string) *Requester {
	if b.Err != nil {
		return b
	}
	if len(p) > 0 {
		b.Paths = append(b.Paths, p)
	}
	return b
}

func (b *Requester) AddPathF(format string, a ...any) *Requester {
	if b.Err != nil {
		return b
	}
	p := fmt.Sprintf(format, a...)
	if len(p) > 0 {
		b.Paths = append(b.Paths, p)
	}
	return b
}

func (b *Requester) AddQuery(key string, value ...string) *Requester {
	if b.Err != nil {
		return b
	}
	for _, v := range value {
		b.Query.Add(key, v)
	}
	return b
}

func (b *Requester) AddQueryAny(key string, value any) *Requester {
	if b.Err != nil {
		return b
	}
	vals, err := toString(value)
	if err != nil {
		b.Err = errors.Join(fmt.Errorf("AddQueryAny err: key %s type %t unsupper", key, value), err)
		return b
	}
	return b.AddQuery(key, vals...)
}

func (b *Requester) SetQuery(key string, value ...string) *Requester {
	if b.Err != nil {
		return b
	}
	b.Header.Del(key)
	return b.AddQuery(key, value...)
}

func (b *Requester) SetQueryAny(key string, value any) *Requester {
	if b.Err != nil {
		return b
	}
	vals, err := toString(value)
	if err != nil {
		b.Err = errors.Join(fmt.Errorf("AddQueryAny err: key %s type %t unsupper", key, value), err)
		return b
	}
	return b.SetQuery(key, vals...)
}

func (b *Requester) SetHeader(key string, value ...string) *Requester {
	if b.Err != nil {
		return b
	}
	b.Header.Del(key)
	return b.AddHeader(key, value...)
}

func (b *Requester) SetHeaderAny(key string, value any) *Requester {
	if b.Err != nil {
		return b
	}
	v, err := toString(value)
	if err != nil {
		b.Err = errors.Join(fmt.Errorf("SetHeaderAny err: key %s type %t unsupper", key, value), err)
		return b
	}
	return b.SetHeader(key, v...)
}

func (b *Requester) AddHeader(key string, value ...string) *Requester {
	if b.Err != nil {
		return b
	}
	for _, s := range value {
		b.Header.Add(key, s)
	}
	return b
}

func (b *Requester) AddHeaderAny(key string, value any) *Requester {
	if b.Err != nil {
		return b
	}
	v, err := toString(value)
	if err != nil {
		b.Err = errors.Join(fmt.Errorf("AddHeaderAny err: key %s type %t unsupper", key, value), err)
		return b
	}
	return b.AddHeader(key, v...)
}

// BindQuery 支持从 map、struct 类型绑定到Query参数
func (b *Requester) BindQuery(a any) *Requester {
	if b.Err != nil {
		return b
	}
	rv := reflect.ValueOf(a)
	if !rv.IsValid() {
		b.Err = errors.New("BindQuery: invalid argument")
		return b
	}
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return b
		}
		return b.BindQuery(rv.Elem().Interface())
	case reflect.Struct:
		tagName := BIND_QUERY_TAG
		if len(b.client.QueryTag) > 0 {
			tagName = b.client.QueryTag
		}
		for i := 0; i < rv.NumField(); i++ {
			ft := rv.Type().Field(i)
			if !ft.IsExported() {
				continue
			}
			tag := ft.Tag.Get(tagName)
			if len(tag) == 0 {
				tag = ft.Name
			}
			if tag == "-" {
				continue
			}
			if ft.Anonymous {
				b.BindQuery(rv.Field(i).Interface())
				continue
			}
			fv := rv.Field(i)
			res, err := toString(fv)
			if err != nil {
				b.Err = errors.Join(errors.New("BindQuery field invalid"), err)
				return b
			}
			b.SetQuery(tag, res...)
		}
		return b
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			if key.Kind() != reflect.String {
				b.Err = errors.New("BindQuery map key It must be string type")
				return b
			}
			res, err := toString(rv.MapIndex(key))
			if err != nil {
				b.Err = errors.Join(errors.New("BindQuery map value invalid"), err)
				return b
			}
			b.SetQuery(key.String(), res...)
		}
		return b
	default:
		b.Err = fmt.Errorf("unsupported type: %T", a)
		return b
	}
}

func (b *Requester) SetBody(contentType string, body io.Reader) *Requester {
	if b.Err != nil {
		return b
	}
	b.Body = body
	b.SetHeader("Content-Type", contentType)
	return b
}

// SetBodyJSON application/json
func (b *Requester) SetBodyJSON(a any) *Requester {
	if b.Err != nil {
		return b
	}
	data, err := json.Marshal(a)
	if err != nil {
		b.Err = err
		return b
	}
	return b.SetBody("application/json", bytes.NewReader(data))
}

// SetBodyForm application/x-www-form-urlencoded
func (b *Requester) SetBodyForm(form url.Values) *Requester {
	if b.Err != nil {
		return b
	}
	return b.SetBody("application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
}

// SetBodyFormMap application/x-www-form-urlencoded
func (b *Requester) SetBodyFormMap(mForm map[string]any) *Requester {
	if b.Err != nil {
		return b
	}
	var form = make(url.Values)
	for k, val := range mForm {
		v, err := toString(val)
		if err != nil {
			b.Err = errors.Join(fmt.Errorf("SetBodyFormMap err: key %s type %t unsupper", k, val), err)
			return b
		}
		form.Del(k)
		for _, s := range v {
			form.Add(k, s)
		}
	}
	return b.SetBodyForm(form)
}

func NewFile(name string, file io.Reader) *File {
	return &File{
		name:   name,
		reader: file,
	}
}

type File struct {
	name   string
	reader io.Reader
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.reader.Read(p)
}

type FileReader interface {
	Name() string
	Read(b []byte) (n int, err error)
}

// SetBodyMultipartForm multipart/form-data，文件的 value 值必须是实现了 FileReader 接口的 *os.File *File或任意的实现接口的类型
func (b *Requester) SetBodyMultipartForm(m map[string]any) *Requester {
	if b.Err != nil {
		return b
	}
	fields := make(map[string][]string)
	files := make(map[string]FileReader)
	for k, val := range m {
		if v, err := toString(val); err == nil {
			fields[k] = v
			continue
		}
		switch v := val.(type) {
		case FileReader:
			files[k] = v
		case *os.File:
			files[k] = v
		case os.File:
			files[k] = &v
		case *File:
			files[k] = v
		case File:
			files[k] = &v
		default:
			b.Err = fmt.Errorf("SetBodyMultipartForm err: key %s type %t unsupper", k, val)
			return b
		}
	}

	if len(files) == 0 {
		buf := bytes.NewBuffer(nil)
		form := multipart.NewWriter(buf)
		for k, vs := range fields {
			for _, v := range vs {
				if err := form.WriteField(k, v); err != nil {
					b.Err = fmt.Errorf("field %s set error: %v", k, err)
					return b
				}
			}
		}
		if err := form.Close(); err != nil {
			b.Err = fmt.Errorf("close form error: %v", err)
			return b
		}
		return b.SetBody(form.FormDataContentType(), buf)
	}
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		for k, vs := range fields {
			for _, v := range vs {
				if err := writer.WriteField(k, v); err != nil {
					b.Err = fmt.Errorf("field %s set error: %v", k, err)
					pw.CloseWithError(err)
					return
				}
			}
		}
		for k, v := range files {
			part, err := writer.CreateFormFile(k, filepath.Base(v.Name()))
			if err != nil {
				b.Err = fmt.Errorf("create form file error: %v", err)
				pw.CloseWithError(err)
				return
			}
			if _, err = io.Copy(part, v); err != nil {
				b.Err = fmt.Errorf("copy file error: %v", err)
				pw.CloseWithError(err)
				return
			}
		}
		if err := writer.Close(); err != nil {
			b.Err = fmt.Errorf("close form error: %v", err)
			pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()
	return b.SetBody(writer.FormDataContentType(), pr)
}

// SetBodyXML application/xml
func (b *Requester) SetBodyXML(a any) *Requester {
	if b.Err != nil {
		return b
	}
	data, err := xml.Marshal(a)
	if err != nil {
		b.Err = err
		return b
	}
	return b.SetBody("application/xml", bytes.NewReader(data))
}

// SetBodyText text/plain
func (b *Requester) SetBodyText(text string) *Requester {
	if b.Err != nil {
		return b
	}
	return b.SetBody("text/plain", strings.NewReader(text))
}

// SetBodyBinary application/octet-stream
func (b *Requester) SetBodyBinary(r io.Reader) *Requester {
	if b.Err != nil {
		return b
	}
	return b.SetBody("application/octet-stream", r)
}

func (b *Requester) Build(ctx context.Context) (*http.Request, error) {
	if b.Err != nil {
		return nil, b.Err
	}
	for _, fn := range b.client.BeforeRequest {
		if err := fn(b); err != nil {
			return nil, err
		}
	}
	urlBuf := make([]byte, len(b.URL))
	copy(urlBuf, b.URL)
	for _, path := range b.Paths {
		if len(path) == 0 {
			continue
		}
		if !bytes.HasSuffix(urlBuf, []byte("/")) && !strings.HasPrefix(path, "/") {
			urlBuf = append(urlBuf, '/')
		}
		urlBuf = append(urlBuf, path...)
	}
	if len(b.Query) > 0 {
		if !bytes.HasSuffix(urlBuf, []byte("?")) {
			urlBuf = append(urlBuf, '?')
		}
		urlBuf = append(urlBuf, b.Query.Encode()...)
	}
	req, err := http.NewRequestWithContext(ctx, b.Method, string(urlBuf), b.Body)
	if err != nil {
		return nil, err
	}
	req.Header = b.Header
	for _, fn := range b.client.AfterRequest {
		if err = fn(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func (b *Requester) Do(ctx context.Context) (resp *http.Response, err error) {
	if b.Err != nil {
		return nil, b.Err
	}
	client := b.client.Client
	if client == nil {
		client = http.DefaultClient
	}
	req, err := b.Build(ctx)
	if err != nil {
		return nil, err
	}
	debug, _ := ctx.Value(ANNE_DEBUG).(bool)
	if debugReq, _ := ctx.Value(ANNE_REQUEST_DEBUG).(bool); debug || debugReq {
		out, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		log.Printf(`[DEBUG] Request "%s"%s`, string(out), "\n")
	}
	if resp, err = client.Do(req); err != nil {
		return nil, err
	}
	if debugResp, _ := ctx.Value(ANNE_RESPONSE_DEBUG).(bool); debug || debugResp {
		out, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		log.Printf(`[DEBUG] Response "%s"%s`, string(out), "\n")
	}
	for _, fn := range b.client.AfterResponse {
		if err = fn(resp); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func (b *Requester) DoDebug(ctx context.Context) (*http.Response, error) {
	return b.Do(context.WithValue(ctx, ANNE_DEBUG, true))
}

type Response struct {
	Err        error
	Status     string
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (r *Response) Validate(fn func(r *Response) error) *Response {
	if r.Err != nil {
		return r
	}
	r.Err = fn(r)
	return r
}

func (r *Response) EqStatusCode(code int) *Response {
	if r.Err != nil {
		return r
	}
	if r.StatusCode != code {
		r.Err = fmt.Errorf("statusCode %d not match statusText %s", r.StatusCode, r.Status)
	}
	return r
}

func (r *Response) JSON(a any) error {
	if r.Err != nil {
		return r.Err
	}
	return json.Unmarshal(r.Body, a)
}

func (r *Response) XML(a any) error {
	if r.Err != nil {
		return r.Err
	}
	return xml.Unmarshal(r.Body, a)
}

func (b *Requester) DoResponse(ctx context.Context) *Response {
	if b.Err != nil {
		return &Response{Err: b.Err}
	}
	resp, err := b.Do(ctx)
	if err != nil {
		return &Response{Err: b.Err}
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return &Response{
		Err:        err,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       data,
	}
}

// Reset reset Err、Method、URL、Header、Body
func (b *Requester) Reset() *Requester {
	b.Err = nil
	b.Method = ""
	b.URL = ""
	b.Header = make(http.Header)
	b.Body = nil
	return b
}
