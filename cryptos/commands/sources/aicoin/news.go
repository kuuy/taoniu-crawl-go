package aicoin

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/crawls/cryptos/commands/sources/aicoin/news"
)

func NewNewsCommand() *cli.Command {
  return &cli.Command{
    Name:  "news",
    Usage: "",
    Subcommands: []*cli.Command{
      news.NewHomeCommand(),
      news.NewCategoriesCommand(),
    },
  }
}
