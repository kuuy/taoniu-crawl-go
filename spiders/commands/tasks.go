package commands

import (
  "github.com/hibiken/asynq"
  "github.com/nats-io/nats.go"
  "gorm.io/gorm"
  "log"
  "taoniu.local/crawls/spiders/queue/asynq/jobs"

  "github.com/urfave/cli/v2"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/repositories"
)

type TasksHandler struct {
  Db         *gorm.DB
  Nats       *nats.Conn
  Asynq      *asynq.Client
  Repository *repositories.TasksRepository
}

func NewTasksCommand() *cli.Command {
  var h TasksHandler
  return &cli.Command{
    Name:  "tasks",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TasksHandler{
        Db:    common.NewDB(),
        Nats:  common.NewNats(),
        Asynq: common.NewAsynqClient(),
      }
      h.Repository = &repositories.TasksRepository{
        Db:    h.Db,
        Nats:  h.Nats,
        Asynq: h.Asynq,
        Job:   &jobs.Tasks{},
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "process",
        Usage: "",
        Action: func(c *cli.Context) error {
          id := c.Args().Get(0)
          if id == "" {
            log.Fatal("id is empty")
            return nil
          }
          if err := h.process(id); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TasksHandler) process(id string) error {
  log.Println("tasks processing...")

  task, err := h.Repository.Get(id)
  if err != nil {
    return err
  }

  return h.Repository.Process(task)
}
