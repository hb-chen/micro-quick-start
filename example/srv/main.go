package main

import (
	"github.com/hb-go/micro-quick-start/example/srv/handler"
	"github.com/hb-go/micro-quick-start/example/srv/subscriber"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-plugins/wrapper/trace/opentracing"

	example "github.com/hb-go/micro-quick-start/example/srv/proto/example"
	tracer "github.com/hb-go/micro-quick-start/pkg/opentracing"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.micro.srv.example"),
		micro.Version("v1"),
	)

	// Initialise service
	metadata := make(map[string]string)
	metadata["key"] = "value1"

	service.Init()
	service.Init(
		micro.Name("go.micro.srv.example"),
		micro.Version("v1"),
		micro.Metadata(metadata),
	)

	// 链路追踪
	t, closer, err := tracer.NewJaegerTracer("example.srv", "127.0.0.1:6831")
	if err != nil {
		log.Fatalf("opentracing tracer create error:%v", err)
	}
	defer closer.Close()
	service.Init(
		micro.WrapHandler(opentracing.NewHandlerWrapper(t)),
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
