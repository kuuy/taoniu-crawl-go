package tasks

import (
  "time"

  "github.com/hibiken/asynq"

  config "taoniu.local/crawls/spiders/config/queue"
  "taoniu.local/crawls/spiders/queue/jobs"
  "taoniu.local/crawls/spiders/repositories"
)

type TasksTask struct {
  Asynq      *asynq.Client
  Job        *jobs.Tasks
  Repository *repositories.TasksRepository
}

func (t *TasksTask) Rescue() error {
  ids := t.Repository.Scan(3)
  for _, id := range ids {
    task, err := t.Job.Process(id)
    if err != nil {
      return err
    }
    t.Asynq.Enqueue(
      task,
      asynq.Queue(config.TASKS),
      asynq.MaxRetry(0),
      asynq.Timeout(5*time.Minute),
    )
  }

  return nil
}
