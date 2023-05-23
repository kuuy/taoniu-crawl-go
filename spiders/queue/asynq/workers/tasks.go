package workers

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/queue/asynq/jobs"
  "taoniu.local/crawls/spiders/repositories"
)

type Tasks struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Nats       *nats.Conn
  Asynq      *asynq.Client
  Ctx        context.Context
  Repository *repositories.TasksRepository
}

func NewTasks() *Tasks {
  h := &Tasks{
    Db:    common.NewDB(),
    Rdb:   common.NewRedis(),
    Nats:  common.NewNats(),
    Asynq: common.NewAsynqClient(),
    Ctx:   context.Background(),
  }
  h.Repository = &repositories.TasksRepository{
    Db:    h.Db,
    Nats:  h.Nats,
    Asynq: h.Asynq,
    Job:   &jobs.Tasks{},
  }
  return h
}

type TasksProcessPayload struct {
  ID string
}

func (h *Tasks) Process(ctx context.Context, t *asynq.Task) error {
  var payload TasksProcessPayload
  json.Unmarshal(t.Payload(), &payload)

  mutex := common.NewMutex(
    h.Rdb,
    h.Ctx,
    fmt.Sprintf("locks:spiders:tasks:process:%s", payload.ID),
  )
  if mutex.Lock(5 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  task, err := h.Repository.Get(payload.ID)
  if err == nil {
    h.Repository.Process(task)
  }

  return nil
}
