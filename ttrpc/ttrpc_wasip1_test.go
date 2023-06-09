//go:build wasip1

package ttrpc_test

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/containerd/ttrpc"
	pb "github.com/stealthrocket/net/ttrpc"
	"github.com/stealthrocket/net/wasip1"
)

type helloService struct {
}

func (s *helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Text: "Hello, " + in.GetName() + "!"}, nil
}

func TestTTRPC(t *testing.T) {
	ctx := context.Background()
	deadline, ok := t.Deadline()
	if ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}
	// First create the listener that the ttrpc server will be using to accept
	// connections using the wasip1 package instead of the standard net package
	// to use WASI socket extensions not available in Go 1.21.
	l, err := wasip1.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	server, err := ttrpc.NewServer()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	pb.RegisterHelloService(server, new(helloService))

	addr := l.Addr().String()
	errs := make(chan error, 1)
	go func() {
		defer close(errs)
		defer server.Shutdown(ctx)
		// Connect to the ttrpc server using the wasip1 dial function to
		// establish a connection using the WASI socket extensions not available
		// in Go 1.21.
		conn, err := wasip1.DialContext(ctx, "tcp", addr)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close()

		client := pb.NewHelloClient(ttrpc.NewClient(conn))
		r, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
		if err != nil {
			errs <- err
			return
		}
		if r.Text != "Hello, World!" {
			t.Errorf("wrong ttrpc response received: %q", r.Text)
		}
	}()

	if err := server.Serve(ctx, l); err != nil {
		if !errors.Is(err, ttrpc.ErrServerClosed) {
			log.Fatalf("%#v\n", err)
		}
	}
	if err := <-errs; err != nil {
		log.Fatal(err)
	}
}
