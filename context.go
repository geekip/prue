package prue

import (
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Path     string
	Data     map[string]interface{}
	Params   map[string]string
	handlers []Handler
	index    int8
}

const abortIndex int8 = math.MaxInt8 >> 1

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Request = req
	ctx.Response = w
	ctx.Path = normalizePath(req.URL.Path)
	ctx.Data = make(map[string]interface{})
	ctx.Params = make(map[string]string)
	ctx.index = -1
	return ctx
}

func releaseContext(ctx *Context) {
	contextPool.Put(ctx)
}

func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if ctx.handlers[ctx.index] == nil {
			continue
		}
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

func (ctx *Context) IsAborted() bool {
	return ctx.index >= abortIndex
}

func (ctx *Context) Abort() {
	ctx.index = abortIndex
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

func (ctx *Context) Cookie(name string) (*http.Cookie, error) {
	return ctx.Request.Cookie(name)
}

func (ctx *Context) Cookies() []*http.Cookie {
	return ctx.Request.Cookies()
}

var templateCache = sync.Map{}

func (ctx *Context) Render(tpl string) {
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

func (ctx *Context) FileAttachment(filepath, filename string) {
	key := "Content-Disposition"
	if isASCII(filename) {
		ctx.Header(key, `attachment; filename="`+quoteEscaper.Replace(filename)+`"`)
	} else {
		ctx.Header(key, `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(ctx.Response, ctx.Request, filepath)
}
