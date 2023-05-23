package nats

import (
  "github.com/nats-io/nats.go"
  "taoniu.local/crawls/spiders/queue/nats/workers"
)

type Workers struct{}

func NewWorkers() *Workers {
  return &Workers{}
}

func (h *Workers) Subscribe(nc *nats.Conn) error {
  workers.NewSources().Subscribe(nc)
  return nil
}
