# micro-quick-start
go-micro微服务快速开始教程

这只是一个简易教程，是结合演示文稿[micro-quick-start.pdf](micro-quick-start.pdf)的代码示例，结合演示文稿可以快速了解微服务与`go-micro`框架，
并可以结合代码示例进行验证测试，文稿中有关于[`micro`自定义](#micro自定义)以及`go-micro`现成组件的使用这里并没有全部示范，详细参考文稿，
本教程每个示例均对应一个`commit`，可以快速查看/比较示例相关代码，包含以下几个阶段：

- 服务创建
	- 通过`micro new`创建服务后进行完善，使模板创建的服务可以运行
- 自定义Server
- 服务筛选
- 自定义Wrapper-监控
- 超时&重试
- 限流&熔断
- 配置-consul
- 链路追踪
- k8s部署

## micro自定义
`micro`最为网关往往需要自定义，`micro`提供了`plugin`的自定义，以下是部分参考，如增加组件`tcp`、`kubernetes`，使用`go-plugins`中已有的`micro/metrics`插件，
以及完全自定义一个`metrics`插件。
```go
package main

import (
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-plugins/micro/metrics"
	"github.com/micro/micro/api"
	"github.com/micro/micro/plugin"
	"github.com/micro/micro/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"

	// tcp transport
	_ "github.com/micro/go-plugins/transport/tcp"

	// k8s registry
	_ "github.com/micro/go-plugins/registry/kubernetes"
)

// Metrics
func init() {
	api.Register(metrics.NewPlugin())
	web.Register(metrics.NewPlugin())
}

// 自定义Metrics
func init() {
	api.Register(plugin.NewPlugin(
		plugin.WithHandler(
			func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					r.Header.Get("")
				})
			}),
	))

	api.Register(plugin.NewPlugin(
		plugin.WithHandler(
			func(h http.Handler) http.Handler {
				md := make(map[string]string)

				opsCounter := prometheus.NewCounterVec(
					prometheus.CounterOpts{
						Namespace: "micro",
						Name:      "request_total",
						Help:      "How many go-micro requests processed, partitioned by method and status",
					},
					[]string{"path", "method", "code"},
				)

				timeCounterSummary := prometheus.NewSummaryVec(
					prometheus.SummaryOpts{
						Namespace: "micro",
						Name:      "upstream_latency_microseconds",
						Help:      "Service backend method request latencies in microseconds",
					},
					[]string{"path", "method"},
				)

				timeCounterHistogram := prometheus.NewHistogramVec(
					prometheus.HistogramOpts{
						Namespace: "micro",
						Name:      "request_duration_seconds",
						Help:      "Service method request time in seconds",
					},
					[]string{"path", "method"},
				)

				reg := prometheus.NewRegistry()
				wrapreg := prometheus.WrapRegistererWith(md, reg)
				wrapreg.MustRegister(
					opsCounter,
					timeCounterSummary,
					timeCounterHistogram,
				)

				prometheus.DefaultGatherer = reg
				prometheus.DefaultRegisterer = wrapreg

				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// 拦截metrics path，默认"/metrics"
					if r.URL.Path == "/metrics" {
						promhttp.Handler().ServeHTTP(w, r)
						return
					}

					path := r.URL.Path
					method := r.Method
					timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
						us := v * 1000000 // make microseconds
						timeCounterSummary.WithLabelValues(path, method).Observe(us)
						timeCounterHistogram.WithLabelValues(path, method).Observe(v)
					}))
					defer timer.ObserveDuration()

					ww := wrapWriter{ResponseWriter: w}
					h.ServeHTTP(&ww, r)
					log.Logf("statusCode: %d, %s", ww.StatusCode, strconv.Itoa(ww.StatusCode))
					opsCounter.WithLabelValues(path, method, strconv.Itoa(ww.StatusCode)).Inc()
				})
			}),
	))

}

type wrapWriter struct {
	StatusCode int
	http.ResponseWriter
}

func (ww *wrapWriter) WriteHeader(statusCode int) {
	log.Logf("statusCode: %d", statusCode)
	ww.StatusCode = statusCode
	ww.ResponseWriter.WriteHeader(statusCode)
}
```