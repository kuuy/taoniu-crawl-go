package queue

import (
  "log"

  "github.com/hibiken/asynq"
  "github.com/urfave/cli/v2"

  "taoniu.local/crawls/spiders/common"
  queue "taoniu.local/crawls/spiders/queue/asynq"
)

type AsynqHandler struct{}

func NewAsynqCommand() *cli.Command {
  var h AsynqHandler
  return &cli.Command{
    Name:  "asynq",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = AsynqHandler{}
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

func (h *AsynqHandler) run() error {
  log.Println("asynq running...")

  worker := common.NewAsynqServer()

  mux := asynq.NewServeMux()
  queue.NewWorkers().Register(mux)
  if err := worker.Run(mux); err != nil {
    return err
  }

  return nil
}
