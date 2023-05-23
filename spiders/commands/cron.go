package commands

import (
  "github.com/hibiken/asynq"
  "github.com/robfig/cron/v3"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "sync"
  "taoniu.local/crawls/spiders/queue/asynq/jobs"
  "taoniu.local/crawls/spiders/repositories"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/tasks"
)

type CronHandler struct {
  Db    *gorm.DB
  Asynq *asynq.Client
}

func NewCronCommand() *cli.Command {
  var h CronHandler
  return &cli.Command{
    Name:  "cron",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CronHandler{
        Db:    common.NewDB(),
        Asynq: common.NewAsynqClient(),
      }
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

func (h *CronHandler) run() error {
  log.Println("cron running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)
  //
  //sources := tasks.SourcesTask{}
  //sources.Repository = &repositories.SourcesRepository{
  //  Db: h.Db,
  //}

  tasks := tasks.TasksTask{
    Asynq: h.Asynq,
    Job:   &jobs.Tasks{},
  }
  tasks.Repository = &repositories.TasksRepository{
    Db: h.Db,
  }

  c := cron.New()
  c.AddFunc("@every 30s", func() {
    //sources.Flush("aicoin-news")
    //sources.Flush("aicoin-news-categories")
  })
  c.AddFunc("@every 1m", func() {
    tasks.Rescue()
  })
  c.Start()

  <-h.wait(wg)

  return nil
}

func (h *CronHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
