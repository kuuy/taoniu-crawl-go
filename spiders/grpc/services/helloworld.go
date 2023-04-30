package services

import (
  "context"
  "log"
  pb "taoniu.local/crawls/spiders/grpc/helloworld"
)

type Helloworld struct {
  pb.UnimplementedGreeterServer
}

func (s *Helloworld) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
  log.Printf("Received: %v", in.GetName())
  return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *Helloworld) SayHelloAgain(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
  log.Printf("Received: %v", in.GetName())
  return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}
