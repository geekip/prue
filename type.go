package prue

import "net/http"

type Application struct {
	Routes Routes
}

// 控制器
type Handler func(ctx *Context)

// 中间件
type Middlewares []func(ctx *Context, next Handler)

// 路由
type Route struct {
	Name        string
	Path        string
	Methods     []string
	Handler     Handler
	Middlewares Middlewares
	SubRoutes   Routes
}

type Routes []Route

// 上下文
type Context struct {
	Data     map[string]interface{}
	Request  *http.Request
	Response http.ResponseWriter
}
