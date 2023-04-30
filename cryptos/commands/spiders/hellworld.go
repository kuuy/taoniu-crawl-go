package spiders

import (
  "context"
  "github.com/urfave/cli/v2"
  "log"
  services "taoniu.local/crawls/cryptos/grpc/services/spiders"
)

type HelloWorldHandler struct {
  Ctx     context.Context
  Service *services.Helloworld
}

func NewHelloworldCommand() *cli.Command {
  var h HelloWorldHandler
  return &cli.Command{
    Name:  "helloworld",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = HelloWorldHandler{
        Ctx: context.Background(),
      }
      h.Service = &services.Helloworld{
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "test",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.test(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *HelloWorldHandler) test() error {
  log.Println("helloworld test processing...")
  err := h.Service.SayHello("fantacy")
  if err != nil {
    return err
  }
  return nil
}
