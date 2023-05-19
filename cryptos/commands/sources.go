package commands

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/crawls/cryptos/commands/sources"
)

func NewSourcesCommand() *cli.Command {
  return &cli.Command{
    Name:  "sources",
    Usage: "",
    Subcommands: []*cli.Command{
      sources.NewAicoinCommand(),
      sources.NewBitpushCommand(),
    },
  }
}
