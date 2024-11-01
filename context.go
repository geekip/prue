package prue

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Method   string
	Path     string
	Pattern  string
	Data     map[string]interface{}
	Params   map[string]string
	Keys     map[string]any
	Next     func()
	mu       sync.RWMutex
}

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return &Context{}
		},
	}
	quoteEscaper  = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
	templateCache = sync.Map{}
)

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Request = r
	ctx.Response = w
	ctx.Method = r.Method
	ctx.Path = r.URL.Path
	ctx.Data = make(map[string]interface{})
	ctx.Params = make(map[string]string)
	ctx.Keys = make(map[string]any)
	return ctx
}

func (ctx *Context) release() {
	contextPool.Put(ctx)
}

func (ctx *Context) Error(message string, statusCode int) {
	http.Error(ctx.Response, message, statusCode)
}

func (ctx *Context) Header(key, val string) {
	if val == "" {
		ctx.Response.Header().Del(key)
		return
	}
	ctx.Response.Header().Set(key, val)
}

func (ctx *Context) GetHeader(key string) string {
	return ctx.Request.Header.Get(key)
}

func (ctx *Context) ContentType() string {
	return filterFlags(ctx.GetHeader("Content-Type"))
}

func (ctx *Context) Set(key string, value any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Keys[key] = value
}

func (ctx *Context) Get(key string) (value any, exists bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	value, exists = ctx.Keys[key]
	return
}

func (ctx *Context) ParseForm() error {
	return ctx.Request.ParseForm()
}

func (ctx *Context) Query(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

func (ctx *Context) FormValue(key string) string {
	return ctx.Request.FormValue(key)
}

func (ctx *Context) FormInt(key string) (int, error) {
	return strconv.Atoi(ctx.FormValue(key))
}

func (ctx *Context) Redirect(url string, code int) {
	http.Redirect(ctx.Response, ctx.Request, url, code)
}

func (ctx *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(ctx.Response, cookie)
}

func (ctx *Context) Cookie(name string) (string, error) {
	cookie, err := ctx.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

func (ctx *Context) Cookies() []*http.Cookie {
	return ctx.Request.Cookies()
}

func (ctx *Context) Html(tpl string) {
	tmpl, ok := templateCache.Load(tpl)
	if !ok {
		var err error
		tmpl, err = template.ParseFiles(tpl)
		if err != nil {
			ctx.Error(err.Error(), http.StatusInternalServerError)
			return
		}
		templateCache.Store(tpl, tmpl)
	}
	if err := tmpl.(*template.Template).Execute(ctx.Response, ctx.Data); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

func (ctx *Context) Json() {
	ctx.Header("Content-Type", "application/json")
	if err := json.NewEncoder(ctx.Response).Encode(ctx.Data); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

func (ctx *Context) Text(text string) {
	ctx.Header("Content-Type", "text/plain")
	if _, err := ctx.Response.Write([]byte(text)); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

func (ctx *Context) GetClientIp() string {
	if ip := ctx.Request.Header.Get("HTTP_CLIENT_IP"); ip != "" {
		return ip
	}
	if ip := ctx.Request.Header.Get("HTTP_X_FORWARDED_FOR"); ip != "" {
		return ip
	}
	return strings.Split(ctx.Request.RemoteAddr, ":")[0]
}

func (c *Context) File(filepath string) {
	http.ServeFile(c.Response, c.Request, filepath)
}

func (c *Context) Files(filepath string, fs http.FileSystem) {
	defer func(path string) {
		c.Request.URL.Path = path
	}(c.Request.URL.Path)

	c.Request.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.Response, c.Request)
}

func (ctx *Context) FileAttachment(filepath, filename string) {
	headerKey := "Content-Disposition"
	commonVal := "attachment; filename"
	if isASCII(filename) {
		ctx.Header(headerKey, commonVal+`="`+quoteEscaper.Replace(filename)+`"`)
	} else {
		ctx.Header(headerKey, commonVal+`*=UTF-8''`+url.QueryEscape(filename))
	}
	ctx.File(filepath)
}
