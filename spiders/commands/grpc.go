package commands

import (
  "fmt"
  "log"
  "net"
  "os"
  pb "taoniu.local/crawls/spiders/grpc/helloworld"
  "taoniu.local/crawls/spiders/grpc/services"

  "github.com/urfave/cli/v2"
  "google.golang.org/grpc"
)

type GrpcHandler struct {
}

func NewGrpcCommand() *cli.Command {
  var h GrpcHandler
  return &cli.Command{
    Name:  "grpc",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = GrpcHandler{}
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *GrpcHandler) run() error {
  log.Println("grpc running...")

  server := grpc.NewServer()

  lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", os.Getenv("SPIDERS_API_PORT")))
  if err != nil {
    log.Fatalf("net.Listen err: %v", err)
  }

  pb.RegisterGreeterServer(server, &services.Helloworld{})

  server.Serve(lis)

  return nil
}
