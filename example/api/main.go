package main

import (
	"github.com/micro/go-micro/util/log"

	"github.com/hb-go/micro-quick-start/example/api/client"
	"github.com/hb-go/micro-quick-start/example/api/handler"
	"github.com/micro/go-micro"

	"github.com/micro/go-micro/api"
	ha "github.com/micro/go-micro/api/handler/api"
	
	example "github.com/hb-go/micro-quick-start/example/api/proto/example"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.micro.api.example"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()
	service.Init(
		micro.Name("go.micro.api.example"),
		micro.Version("latest"),

		// create wrap for the Example srv client
		micro.WrapHandler(client.ExampleWrapper(service)),
	)

	// Register Handler
	example.RegisterExampleHandler(
		service.Server(),
		new(handler.Example),
		api.WithEndpoint(&api.Endpoint{
			// The RPC method
			Name: "Example.Call",
			// The HTTP paths. This can be a POSIX regex
			Path: []string{"/example/call"},
			// The HTTP Methods for this endpoint
			Method:  []string{"GET", "POST"},
			Handler: ha.Handler,
		}),
	)

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
