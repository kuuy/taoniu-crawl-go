package spiders

import (
  "context"
  "fmt"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "log"
  "os"
  pb "taoniu.local/crawls/cryptos/grpc/spiders/helloworld"
)

type Helloworld struct {
  Ctx           context.Context
  GreeterClient pb.GreeterClient
}

func (s *Helloworld) Client() pb.GreeterClient {
  if s.GreeterClient == nil {
    conn, err := grpc.Dial(
      fmt.Sprintf("127.0.0.1:%v", os.Getenv("SPIDERS_API_PORT")),
      grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
      panic(err.Error())
    }
    s.GreeterClient = pb.NewGreeterClient(conn)
  }
  return s.GreeterClient
}

func (s *Helloworld) SayHello(name string) error {
  r, err := s.Client().SayHello(s.Ctx, &pb.HelloRequest{Name: name})
  if err != nil {
    return err
  }
  log.Printf("Greeting: %s", r.GetMessage())
  return nil
}
