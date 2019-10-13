package handler

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/hb-go/micro-quick-start/example/api/client"
	api "github.com/micro/go-micro/api/proto"
	mc "github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/selector"
	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/util/log"

	example "github.com/hb-go/micro-quick-start/example/srv/proto/example"
)

type Example struct{}

func extractValue(pair *api.Pair) string {
	if pair == nil {
		return ""
	}
	if len(pair.Values) == 0 {
		return ""
	}
	return pair.Values[0]
}

// Example.Call is called by the API as /example/call with post body {"name": "foo"}
func (e *Example) Call(ctx context.Context, req *api.Request, rsp *api.Response) error {
	log.Log("Received Example.Call request")

	// extract the client from the context
	exampleClient, ok := client.ExampleFromContext(ctx)
	if !ok {
		return errors.InternalServerError("go.micro.api.example.example.call", "example client not found")
	}

	// make request
	responses := []*example.Response{}
	mtx := sync.RWMutex{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		response, err := exampleClient.Call(
			ctx,
			&example.Request{
				Name: extractValue(req.Post["name"]) + " 1",
			},
		)
		if err != nil {
			log.Logf("go.micro.api.example.example.call", err.Error())
		} else {
			mtx.Lock()
			responses = append(responses, response)
			mtx.Unlock()
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		response1, err := exampleClient.Call(
			ctx,
			&example.Request{
				Name: extractValue(req.Post["name"]) + " 2",
			},
			mc.WithSelectOption(
				selector.WithFilter(
					selector.FilterLabel("key", "value1"),
				),
			),
		)
		if err != nil {
			log.Logf("go.micro.api.example.example.call", err.Error())
		} else {
			mtx.Lock()
			responses = append(responses, response1)
			mtx.Unlock()
		}
		wg.Done()
	}()
	wg.Wait()

	b, _ := json.Marshal(responses)

	rsp.StatusCode = 200
	rsp.Body = string(b)

	return nil
}
