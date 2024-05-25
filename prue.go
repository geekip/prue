package prue

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// 启动服务
func (app *Application) Listen(addr string) {
	router := mux.NewRouter()
	mountRoutes(router, app.Routes)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("Server is listening on %s\n", addr)

	if err := http.Serve(listener, router); err != nil {
		log.Fatal(err)
	}
}

// 实例化
func Init(routes Routes) *Application {
	return &Application{Routes: routes}
}
