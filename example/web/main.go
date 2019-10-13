package main

import (
	"github.com/micro/go-micro/util/log"
	"net/http"

	"github.com/hb-go/micro-quick-start/example/web/handler"
	"github.com/micro/go-micro/web"
)

func main() {
	// create new web service
	service := web.NewService(
		web.Name("go.micro.web.example"),
		web.Version("latest"),
	)

	// initialise service
	if err := service.Init(); err != nil {
		log.Fatal(err)
	}

	// register html handler
	service.Handle("/", http.FileServer(http.Dir("html")))

	// register call handler
	service.HandleFunc("/example/call", handler.ExampleCall)

	// run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
