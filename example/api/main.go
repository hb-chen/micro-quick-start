package main

import (
	"context"
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

	"github.com/afex/hystrix-go/hystrix"
	"github.com/micro/go-plugins/wrapper/ratelimiter/uber"

	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/consul"

	tracer "github.com/hb-go/micro-quick-start/pkg/opentracing"
	"github.com/micro/go-plugins/wrapper/trace/opentracing"

	example "github.com/hb-go/micro-quick-start/example/api/proto/example"
)

func main() {
	w := conf()
	if w != nil {
		defer func() {
			w.Stop()
		}()
	}

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

	// 限流&熔断
	service.Init(
		micro.WrapClient(ratelimit.NewClientWrapper(10)),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(10)),
		micro.WrapClient(NewClientWrapper()),
	)

	// 链路追踪
	t, closer, err := tracer.NewJaegerTracer("example.api", "127.0.0.1:6831")
	if err != nil {
		log.Fatalf("opentracing tracer create error:%v", err)
	}
	defer closer.Close()
	service.Init(
		micro.WrapCall(opentracing.NewCallWrapper(t)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(t)),
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

func conf() config.Watcher {
	consulSource := consul.NewSource(
		consul.WithPrefix("/micro/config"),
		consul.StripPrefix(true),
	)
	// 加载配置
	err := config.Load(consulSource)
	if err != nil {
		log.Logf("config load error: %v", err)
	} else {
		log.Logf("config data: %v", string(config.Bytes()))
	}

	// 监控动态配置
	w, err := config.Watch("key1")
	if err == nil {
		go func() {
			for {
				if v, err := w.Next(); err == nil {
					log.Logf("config : %v", string(v.Bytes()))
				} else {
					log.Log("config error: %v", err)
					return
				}
			}
		}()

		return w
	} else {
		log.Logf("config watch error: %v", err)
	}

	return nil
}

type clientWrapper struct {
	mc.Client
}

func (c *clientWrapper) Call(ctx context.Context, req mc.Request, rsp interface{}, opts ...mc.CallOption) error {
	return hystrix.Do(req.Service()+"."+req.Endpoint(), func() error {
		if cir, ok, _ := hystrix.GetCircuit(req.Service() + "." + req.Endpoint()); ok {
			log.Logf("circuit: %v %v", cir.Name, cir.AllowRequest())
		} else {
			log.Logf("circuit: %v %v", cir.Name, cir.AllowRequest())
		}

		return c.Client.Call(ctx, req, rsp, opts...)
	}, func(err error) error {
		log.Logf("fallback: %v", err)
		return err
	})
}

// NewClientWrapper returns a hystrix client Wrapper.
func NewClientWrapper() mc.Wrapper {
	return func(c mc.Client) mc.Client {
		return &clientWrapper{c}
	}
}
