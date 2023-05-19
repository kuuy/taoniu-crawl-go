package sources

import (
  "github.com/urfave/cli/v2"
  "taoniu.local/crawls/cryptos/commands/sources/bitpush"
)

func NewBitpushCommand() *cli.Command {
  return &cli.Command{
    Name:  "bitpush",
    Usage: "",
    Subcommands: []*cli.Command{
      bitpush.NewHomeCommand(),
    },
  }
}
