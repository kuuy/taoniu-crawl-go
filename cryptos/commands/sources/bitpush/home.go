package bitpush

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  services "taoniu.local/crawls/cryptos/grpc/services/spiders"
  repositories "taoniu.local/crawls/cryptos/repositories/spiders"
)

type HomeHandler struct {
  Ctx        context.Context
  Repository *repositories.SourcesRepository
}

func NewHomeCommand() *cli.Command {
  var h HomeHandler
  return &cli.Command{
    Name:  "home",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = HomeHandler{
        Ctx: context.Background(),
      }
      h.Repository = &repositories.SourcesRepository{}
      h.Repository.Service = &services.Sources{
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "save",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.save(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *HomeHandler) save() error {
  log.Println("sources save processing...")
  return nil
}
