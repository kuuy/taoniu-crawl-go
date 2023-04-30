package spiders

import (
  "fmt"
  "os"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
)

type SourcesRepository struct{}

func (r *SourcesRepository) Add() error {
  conn, err := grpc.Dial(
    fmt.Sprintf("127.0.0.1:%v", os.Getenv("SPIDERS_API_PORT")),
    grpc.WithTransportCredentials(insecure.NewCredentials()),
  )
  if err != nil {
    return err
  }
  defer conn.Close()

  return nil
}
