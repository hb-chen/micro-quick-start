package handler

import (
	"context"
	"time"

	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/util/log"

	example "github.com/hb-go/micro-quick-start/example/srv/proto/example"
)

type Example struct{}

var count int

// Call is a single request handler called via client.Call or the generated client code
func (e *Example) Call(ctx context.Context, req *example.Request, rsp *example.Response) error {
	log.Log("Received Example.Call request")
	count++
	log.Logf("Received Example.Call request, count: %d", count)

	if count%4 == 0 {
		log.Log("Received Example.Call request 错误")
		return errors.New("go.micro.srv.example.call", "错误", 500)
	}

	log.Log("Received Example.Call request 超时")
	time.Sleep(time.Millisecond * 10)

	rsp.Msg = "Hello " + req.Name + " v1 value1"
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Example) Stream(ctx context.Context, req *example.StreamingRequest, stream example.Example_StreamStream) error {
	log.Logf("Received Example.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Logf("Responding: %d", i)
		if err := stream.Send(&example.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Example) PingPong(ctx context.Context, stream example.Example_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Logf("Got ping %v", req.Stroke)
		if err := stream.Send(&example.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}
