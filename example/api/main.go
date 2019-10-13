package main

import (
	"net"
	"net/http"
	"time"

	"github.com/hb-go/micro-quick-start/example/api/client"
	"github.com/hb-go/micro-quick-start/example/api/handler"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/api"
	ha "github.com/micro/go-micro/api/handler/api"
	mc "github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/selector"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

	// 筛选器
	service.Init(
		func(options *micro.Options) {
			options.Client.Init(func(options *mc.Options) {
				options.CallOptions.SelectOptions =
					append(options.CallOptions.SelectOptions,
						selector.WithFilter(selector.FilterVersion("v1")))
			})
		},
	)

	// 监控
	service.Init(
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
	)
	go func() {
		ls, err := net.Listen("tcp", ":9091")
		if err == nil {
			err := http.Serve(ls, promhttp.Handler())
			if err != nil {
			}
			panic(err)
		} else {
		}
		panic(err)
	}()

	// 重试&超时
	service.Init(
		func(o *micro.Options) {
			o.Client.Init(
				mc.Retries(3),
				mc.RequestTimeout(time.Millisecond*100),
			)
		},
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
