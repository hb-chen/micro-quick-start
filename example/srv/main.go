package main

import (
	"github.com/hb-go/micro-quick-start/example/srv/handler"
	"github.com/hb-go/micro-quick-start/example/srv/subscriber"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/util/log"

	example "github.com/hb-go/micro-quick-start/example/srv/proto/example"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.micro.srv.example"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()
	service.Init(
		micro.Name("go.micro.srv.example"),
		micro.Version("latest"),
	)

	// Register Handler
	example.RegisterExampleHandler(service.Server(), new(handler.Example))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("go.micro.srv.example", service.Server(), new(subscriber.Example))

	// Register Function as Subscriber
	micro.RegisterSubscriber("go.micro.srv.example", service.Server(), subscriber.Handler)

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
