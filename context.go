package prue

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// 定义上下文结构体
type Context struct {
	Data     map[string]interface{}
	Request  *http.Request
	Response http.ResponseWriter
}

// 创建上下文实例
func makeHandler(handlerFunc func(*Context)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := &Context{
			Data:     make(map[string]interface{}),
			Request:  request,
			Response: response,
		}
		handlerFunc(ctx)
	}
}

// 根据key获取数据
func (ctx *Context) GetData(key string) interface{} {
	return ctx.Data[key]
}

// 设置Header
func (ctx *Context) Header(key, val string) {
	ctx.Response.Header().Set(key, val)
}

// 渲染页面模板
func (ctx *Context) Render(tpl string) {
	tmpl := template.Must(template.ParseFiles(tpl))
	if err := tmpl.Execute(ctx.Response, ctx.Data); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

// 输出json
func (ctx *Context) Json() {
	ctx.Header("Content-Type", "application/json")
	if err := json.NewEncoder(ctx.Response).Encode(ctx.Data); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

// 输出text
func (ctx *Context) Text(text string) {
	ctx.Header("Content-Type", "text/plain")
	_, err := ctx.Response.Write([]byte(text))
	if err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

// 获取请求参数
func (ctx *Context) Vars() map[string]string {
	return mux.Vars(ctx.Request)
}

// 获取当前路由
func (ctx *Context) CurrentRoute() *mux.Route {
	return mux.CurrentRoute(ctx.Request)
}

// 响应错误
func (ctx *Context) Error(error string, code int) {
	http.Error(ctx.Response, error, code)
}

// 解析表单
func (ctx *Context) ParseForm() error {
	return ctx.Request.ParseForm()
}

// 获取URL参数
func (ctx *Context) QueryParam(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

// 获取表单值
func (ctx *Context) FormValue(key string) string {
	return ctx.Request.FormValue(key)
}

func (ctx *Context) FormInt(key string) (int, error) {
	val := ctx.FormValue(key)
	return strconv.Atoi(val)
}

// 跳转
func (ctx *Context) Redirect(url string, code int) {
	http.Redirect(ctx.Response, ctx.Request, url, code)
	// http.Redirect(ctx.Response, ctx.Request, url, http.StatusFound)
}

// 获取客户端IP
func (ctx *Context) GetClientIp() string {
	strIp := ctx.Request.Header.Get("HTTP_CLIENT_IP")
	if strIp != "" {
		return strIp
	}
	strIp = ctx.Request.Header.Get("HTTP_X_FORWARDED_FOR")
	if strIp != "" {
		return strIp
	}
	strIp = ctx.Request.Header.Get("REMOTE_ADDR")
	if strIp != "" {
		return strIp
	}
	return strings.Split(ctx.Request.RemoteAddr, ":")[0]
}
