package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/crawls/spiders/commands/queue"
)

func NewQueueCommand() *cli.Command {
  return &cli.Command{
    Name:  "queue",
    Usage: "",
    Subcommands: []*cli.Command{
      queue.NewAsynqCommand(),
      queue.NewNatsCommand(),
    },
  }
}
