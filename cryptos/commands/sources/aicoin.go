package sources

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/crawls/cryptos/commands/sources/aicoin"
)

func NewAicoinCommand() *cli.Command {
  return &cli.Command{
    Name:  "aicoin",
    Usage: "",
    Subcommands: []*cli.Command{
      aicoin.NewNewsCommand(),
    },
  }
}
