package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/crawls/cryptos/commands/spiders"
)

func NewSpidersCommand() *cli.Command {
  return &cli.Command{
    Name:  "spiders",
    Usage: "",
    Subcommands: []*cli.Command{
      spiders.NewSourcesCommand(),
    },
  }
}
