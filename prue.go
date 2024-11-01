package prue

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type (
	Handler func(*Context)
	Mux     struct {
		prefix                  string
		methods                 []string
		node                    *node
		middlewares             []Handler
		notFoundHandler         Handler
		methodNotAllowedHandler Handler
		internalErrorHandler    func(*Context, interface{})
		panicHandler            func(error)
	}
)

var (
	DefaultNotFoundHandler = func(c *Context) {
		c.Error("404 page not found", http.StatusNotFound)
	}
	DefaultMethodNotAllowedHandler = func(c *Context) {
		c.Error("405 method not allowed", http.StatusMethodNotAllowed)
	}
	DefaultInternalErrorHandler = func(c *Context, err interface{}) {
		c.Error("500 internal server error", http.StatusInternalServerError)
	}
	DefaultPanicHandler = func(err error) { panic(err) }
	errorHandleText     = errors.New("prue mux Handle error")
	errorMiddlewareText = errors.New("prue mux unkown middleware")
	errorMethodText     = errors.New("prue mux unkown http method")
)

func New() *Mux {
	return &Mux{
		node:                    newNode(""),
		notFoundHandler:         DefaultNotFoundHandler,
		methodNotAllowedHandler: DefaultMethodNotAllowedHandler,
		internalErrorHandler:    DefaultInternalErrorHandler,
		panicHandler:            DefaultPanicHandler,
	}
}

func (m *Mux) Group(pattern string) *Mux {
	return &Mux{
		prefix:                  pathJoin(m.prefix, pattern),
		node:                    m.node,
		middlewares:             m.middlewares,
		notFoundHandler:         m.notFoundHandler,
		methodNotAllowedHandler: m.methodNotAllowedHandler,
		internalErrorHandler:    m.internalErrorHandler,
		panicHandler:            m.panicHandler,
	}
}

func (m *Mux) Use(middlewares ...Handler) *Mux {
	if len(middlewares) == 0 {
		m.panicHandler(errorMiddlewareText)
	}
	m.middlewares = append(m.middlewares, middlewares...)
	return m
}

func (m *Mux) Method(methods ...string) *Mux {
	if len(methods) == 0 {
		m.panicHandler(errorMethodText)
	}
	m.methods = append(m.methods, methods...)
	return m
}

func (m *Mux) Handle(pattern string, handler Handler) *Mux {
	fullPattern := pathJoin(m.prefix, pattern)
	if len(m.methods) == 0 {
		m.methods = append(m.methods, wildcardPrefix)
	}
	for _, method := range m.methods {
		node := m.node.add(strings.ToUpper(method), fullPattern, handler, m.middlewares)
		if node == nil {
			m.panicHandler(errorHandleText)
		}
	}
	m.methods = nil
	return m
}

func (m *Mux) GET(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodGet).Handle(pattern, handler)
}

func (m *Mux) POST(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodPost).Handle(pattern, handler)
}

func (m *Mux) PUT(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodPut).Handle(pattern, handler)
}

func (m *Mux) DELETE(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodDelete).Handle(pattern, handler)
}

func (m *Mux) PATCH(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodPatch).Handle(pattern, handler)
}

func (m *Mux) OPTIONS(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodOptions).Handle(pattern, handler)
}

func (m *Mux) HEAD(pattern string, handler Handler) *Mux {
	return m.Method(http.MethodHead).Handle(pattern, handler)
}

func (m *Mux) NotFoundHandler(handler Handler) *Mux {
	m.notFoundHandler = handler
	return m
}

func (m *Mux) InternalErrorHandler(handler func(*Context, interface{})) *Mux {
	m.internalErrorHandler = handler
	return m
}

func (m *Mux) MethodNotAllowedHandler(handler Handler) *Mux {
	m.methodNotAllowedHandler = handler
	return m
}

func (m *Mux) PanicHandler(handler func(error)) *Mux {
	m.panicHandler = handler
	return m
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r)
	defer func() {
		if err := recover(); err != nil {
			m.internalErrorHandler(ctx, err)
			ctx.release()
		}
	}()

	var handler Handler
	node := m.node.find(ctx.Method, ctx.Path)
	if node == nil {
		handler = m.notFoundHandler
	} else {
		ctx.Params = node.params
		// ctx.Pattern = node.pattern
		handler = node.handler
		if handler == nil {
			handler = m.methodNotAllowedHandler
		}
		if len(node.middlewares) > 0 {
			handler = m.withMiddleware(node.middlewares, handler)
		}
	}
	handler(ctx)
}

// Apply Middleware
func (m *Mux) withMiddleware(middlewares []Handler, handler Handler) Handler {
	count := len(middlewares)
	// Insert the handler at the end of the middleware
	middlewares = append(middlewares, handler)
	for i := count; i >= 0; i-- {
		next := handler
		handler = func(c *Context) {
			if i < count {
				// Pass next to the middleware
				c.Next = func() { next(c) }
			} else {
				// Remove Next from the handler
				c.Next = nil
			}
			middlewares[i](c)
		}
	}
	return handler
}

func (m *Mux) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		debugPrintError(err)
		return err
	}
	debugPrint("Server is listening on %s\n", addr)
	return m.runWithListener(listener)
}

func (m *Mux) RunListener(listener net.Listener) error {
	debugPrint("Server is listening on %s\n", listener.Addr().String())
	return m.runWithListener(listener)
}

func (m *Mux) runWithListener(listener net.Listener) error {
	defer func() {
		if r := recover(); r != nil {
			debugPrintError(fmt.Errorf("server panic: %v", r))
		}
	}()
	err := http.Serve(listener, m)
	if err != nil {
		debugPrintError(err)
	}
	return err
}
