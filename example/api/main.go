package main

import (
	"github.com/micro/go-micro/util/log"

	"github.com/micro/go-micro"
	"github.com/hb-go/micro-quick-start/example/api/handler"
	"github.com/hb-go/micro-quick-start/example/api/client"

	example "github.com/hb-go/micro-quick-start/example/api/proto/example"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.micro.api.example"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init(
		// create wrap for the Example srv client
		micro.WrapHandler(client.ExampleWrapper(service)),
	)

	// Register Handler
	example.RegisterExampleHandler(service.Server(), new(handler.Example))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
