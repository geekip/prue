package prue

import (
	"fmt"
	"net"
	"net/http"
	"sync"
)

const Version = "v1.0.0"

type application struct {
	Router Router
}

var (
	once sync.Once
	app  *application
)

func New(router Router) *application {
	once.Do(func() {
		app = &application{Router: router}
	})
	return app
}

func (a *application) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		debugPrintError(err)
		return err
	}
	debugPrint("Server is listening on %s\n", addr)
	return a.runWithListener(listener)
}

func (a *application) RunListener(listener net.Listener) error {
	debugPrint("Server is listening on %s\n", listener.Addr().String())
	return a.runWithListener(listener)
}

func (a *application) runWithListener(listener net.Listener) error {
	defer func() {
		if r := recover(); r != nil {
			debugPrintError(fmt.Errorf("server panic: %v", r))
		}
	}()

	err := http.Serve(listener, &a.Router)
	if err != nil {
		debugPrintError(err)
	}
	return err
}
