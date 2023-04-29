package commands

import (
  "gorm.io/gorm"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/models"
  "taoniu.local/crawls/spiders/repositories"
)

type SourcesHandler struct {
  Db         *gorm.DB
  Repository *repositories.SourcesRepository
}

func NewSourcesCommand() *cli.Command {
  var h SourcesHandler
  return &cli.Command{
    Name:  "sources",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SourcesHandler{
        Db: common.NewDB(),
      }
      h.Repository = &repositories.SourcesRepository{
        Db: h.Db,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "add",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.add(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "crawl",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.crawl(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *SourcesHandler) add() error {
  log.Println("sources add processing...")
  parentId := ""
  name := "资讯（AICOIN）"
  slug := "aicoin-news"
  url := "https://www.aicoin.com/news/all"
  headers := map[string]string{
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
  }
  extractRules := make(map[string]*repositories.ExtractRules)
  extractRules["categories"] = &repositories.ExtractRules{
    Container: &repositories.HtmlExtractNode{
      Selector: "#news_tabs",
    },
    List: &repositories.HtmlExtractNode{
      Selector: "ul.nav li a",
    },
    Fields: []*repositories.HtmlExtractField{
      {
        Name: "title",
        Node: &repositories.HtmlExtractNode{},
      },
      {
        Name: "link",
        Node: &repositories.HtmlExtractNode{
          Attr:  "href",
          Index: 0,
        },
        RegexReplace: []*repositories.RegexReplace{
          {
            Pattern: `/news/([^/]+)`,
            Value:   "$1",
          },
        },
      },
    },
  }
  source := &models.Source{
    Url:          url,
    Headers:      h.Repository.JSONMap(headers),
    UseProxy:     false,
    Timeout:      10,
    ExtractRules: h.Repository.JSONMap(extractRules),
  }

  return h.Repository.Add(parentId, name, slug, source)
}

func (h *SourcesHandler) crawl() error {
  log.Println("sources crawl processing...")
  source, err := h.Repository.Get("aicoin-news")
  if err != nil {
    return err
  }
  err = h.Repository.Crawl(source)
  if err != nil {
    return err
  }
  return nil
}
