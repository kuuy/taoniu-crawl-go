package jobs

import (
  "encoding/json"
  "github.com/hibiken/asynq"
)

type Tasks struct{}

type TasksProcessPayload struct {
  ID string
}

func (h *Tasks) Process(id string) (*asynq.Task, error) {
  payload, err := json.Marshal(TasksProcessPayload{id})
  if err != nil {
    return nil, err
  }
  return asynq.NewTask("tasks:process", payload), nil
}
