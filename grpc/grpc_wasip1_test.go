//go:build wasip1

package grpc_test

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/stealthrocket/net/grpc"
	"github.com/stealthrocket/net/wasip1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type helloService struct {
	pb.UnimplementedHelloServiceServer
}

func (s *helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Text: "Hello, " + in.GetName() + "!"}, nil
}

func TestGRPC(t *testing.T) {
	// First create the listener that the gRPC server will be using to accept
	// connections using the wasip1 package instead of the standard net package
	// to use WASI socket extensions not available in Go 1.21.
	l, err := wasip1.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	server := grpc.NewServer()
	pb.RegisterHelloServiceServer(server, new(helloService))

	addr := l.Addr().String()
	errs := make(chan error, 1)
	go func() {
		defer close(errs)
		defer server.GracefulStop()

		// When opening a gRPC connection, we must configure the dialer to use
		// the wasip1 package instead of the default.
		conn, err := grpc.Dial(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
				return wasip1.DialContext(ctx, "tcp", address)
			}),
		)
		if err != nil {
			errs <- err
			return
		}
		defer conn.Close()
		client := pb.NewHelloServiceClient(conn)

		ctx := context.Background()
		deadline, ok := t.Deadline()
		if ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithDeadline(ctx, deadline)
			defer cancel()
		}

		r, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
		if err != nil {
			errs <- err
			return
		}
		if r.Text != "Hello, World!" {
			t.Errorf("wrong gRPC response received: %q", r.Text)
		}
	}()

	if err := server.Serve(l); err != nil {
		log.Fatalf("%#v\n", err)
	}
	if err := <-errs; err != nil {
		log.Fatal(err)
	}
}
