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
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.flush(); err != nil {
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
  extractRules["news-list"] = &repositories.ExtractRules{
    Container: &repositories.HtmlExtractNode{
      Selector: "#news_tabs",
    },
    List: &repositories.HtmlExtractNode{
      Selector: "ul._2VG1sacuKAPPtahDo9EmQd li.clearfix",
    },
    Fields: []*repositories.HtmlExtractField{
      {
        Name: "title",
        Node: &repositories.HtmlExtractNode{
          Selector: "div.news-title h3 a",
        },
      },
      {
        Name: "link",
        Node: &repositories.HtmlExtractNode{
          Selector: "div.news-title h3 a",
          Attr:     "href",
        },
        RegexReplace: []*repositories.RegexReplace{
          {
            Pattern: `/article/([^/]+)`,
            Value:   "$1",
          },
        },
      },
      {
        Name: "source",
        Node: &repositories.HtmlExtractNode{
          Selector: "div.news-info span.category",
        },
      },
      {
        Name: "published-time",
        Node: &repositories.HtmlExtractNode{
          Selector: "div.news-info span.news-published-time",
        },
      },
    },
  }
  extractRules["top-list"] = &repositories.ExtractRules{
    Container: &repositories.HtmlExtractNode{
      Selector: "ul.top-list",
    },
    List: &repositories.HtmlExtractNode{
      Selector: "li h3 a",
    },
    Fields: []*repositories.HtmlExtractField{
      {
        Name: "title",
        Node: &repositories.HtmlExtractNode{},
      },
      {
        Name: "link",
        Node: &repositories.HtmlExtractNode{
          Attr: "href",
        },
        RegexReplace: []*repositories.RegexReplace{
          {
            Pattern: `/article/([^/]+)`,
            Value:   "$1",
          },
        },
      },
    },
  }
  extractRules["hot-list"] = &repositories.ExtractRules{
    Container: &repositories.HtmlExtractNode{
      Selector: "div.hot-list",
    },
    List: &repositories.HtmlExtractNode{
      Selector: "div.news-detail",
    },
    Fields: []*repositories.HtmlExtractField{
      {
        Name: "title",
        Node: &repositories.HtmlExtractNode{
          Selector: "div.news-title a",
        },
      },
      {
        Name: "link",
        Node: &repositories.HtmlExtractNode{
          Selector: "div.news-title a",
          Attr:     "href",
        },
        RegexReplace: []*repositories.RegexReplace{
          {
            Pattern: `/article/([^/]+).html`,
            Value:   "$1",
          },
        },
      },
      {
        Name: "category",
        Node: &repositories.HtmlExtractNode{
          Selector: "span.news-category",
        },
      },
      {
        Name: "published-time",
        Node: &repositories.HtmlExtractNode{
          Selector: "span.news-published-time",
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

func (h *SourcesHandler) flush() error {
  log.Println("sources flush processing...")
  source, err := h.Repository.Get("aicoin-news")
  if err != nil {
    return err
  }
  err = h.Repository.Flush(source)
  if err != nil {
    return err
  }
  return nil
}
