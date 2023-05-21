package queue

import (
  "github.com/hibiken/asynq"
  "taoniu.local/crawls/spiders/queue/workers"
)

type Workers struct{}

func NewWorkers() *Workers {
  return &Workers{}
}

func (h *Workers) Register(mux *asynq.ServeMux) error {
  mux.HandleFunc("tasks:process", workers.NewTasks().Process)
  return nil
}
