package spiders

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  services "taoniu.local/crawls/cryptos/grpc/services/spiders"
  repositories "taoniu.local/crawls/cryptos/repositories/spiders"
)

type SourcesHandler struct {
  Ctx        context.Context
  Repository *repositories.SourcesRepository
}

func NewSourcesCommand() *cli.Command {
  var h SourcesHandler
  return &cli.Command{
    Name:  "sources",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SourcesHandler{
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

func (h *SourcesHandler) save() error {
  log.Println("sources save processing...")
  parentId := ""
  name := "资讯（AICOIN）"
  slug := "aicoin-news"
  url := "https://www.aicoin.com/news/all"
  headers := map[string]string{
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
  }
  extractRules := make(map[string]interface{})

  var rules map[string]interface{}
  var html map[string]interface{}
  var container map[string]interface{}
  var list map[string]interface{}
  var node map[string]interface{}
  var fields []interface{}
  var field map[string]interface{}
  var regexReplace []map[string]string

  html = make(map[string]interface{})

  container = make(map[string]interface{})
  container["selector"] = "#news_tabs"

  list = make(map[string]interface{})
  list["selector"] = "ul.nav li[role='presentation']"

  html["container"] = container
  html["list"] = list

  fields = make([]interface{}, 0)

  node = make(map[string]interface{})
  node["selector"] = "a"
  field = make(map[string]interface{})
  field["name"] = "title"
  field["node"] = node
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "a"
  node["attr"] = "href"
  field = make(map[string]interface{})
  field["name"] = "slug"
  field["node"] = node
  regexReplace = make([]map[string]string, 0)
  regexReplace = append(regexReplace, map[string]string{
    "pattern": "/news/([^/]+)",
    "value":   "$1",
  })
  field["regex_replace"] = regexReplace
  fields = append(fields, field)

  html["fields"] = fields

  rules = make(map[string]interface{})
  rules["html"] = html
  extractRules["categories"] = rules

  html = make(map[string]interface{})

  container = make(map[string]interface{})
  container["selector"] = "#news_tabs"

  list = make(map[string]interface{})
  list["selector"] = "ul._2VG1sacuKAPPtahDo9EmQd li.clearfix"

  html["container"] = container
  html["list"] = list

  fields = make([]interface{}, 0)

  node = make(map[string]interface{})
  node["selector"] = "div.news-title h3 a"
  field = make(map[string]interface{})
  field["name"] = "title"
  field["node"] = node
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "div.news-title h3 a"
  node["attr"] = "href"
  field = make(map[string]interface{})
  field["name"] = "id"
  field["node"] = node
  regexReplace = make([]map[string]string, 0)
  regexReplace = append(regexReplace, map[string]string{
    "pattern": "/article/([^/]+).html",
    "value":   "$1",
  })
  field["regex_replace"] = regexReplace
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "div.news-info span.category"
  field = make(map[string]interface{})
  field["name"] = "source"
  field["node"] = node
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "div.news-info span.news-published-time"
  field = make(map[string]interface{})
  field["name"] = "published-time"
  field["node"] = node
  fields = append(fields, field)

  html["fields"] = fields

  rules = make(map[string]interface{})
  rules["html"] = html
  extractRules["news-list"] = rules

  html = make(map[string]interface{})

  container = make(map[string]interface{})
  container["selector"] = "ul.top-list"

  list = make(map[string]interface{})
  list["selector"] = "li h3 a"

  html["container"] = container
  html["list"] = list

  fields = make([]interface{}, 0)

  node = make(map[string]interface{})
  field = make(map[string]interface{})
  field["name"] = "title"
  field["node"] = node
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["attr"] = "href"
  field = make(map[string]interface{})
  field["name"] = "id"
  field["node"] = node
  regexReplace = make([]map[string]string, 0)
  regexReplace = append(regexReplace, map[string]string{
    "pattern": "/article/([^/]+).html",
    "value":   "$1",
  })
  field["regex_replace"] = regexReplace
  fields = append(fields, field)

  html["fields"] = fields

  rules = make(map[string]interface{})
  rules["html"] = html
  extractRules["top-list"] = rules

  html = make(map[string]interface{})

  container = make(map[string]interface{})
  container["selector"] = "div.hot-list"

  list = make(map[string]interface{})
  list["selector"] = "div.news-detail"

  html["container"] = container
  html["list"] = list

  fields = make([]interface{}, 0)

  node = make(map[string]interface{})
  node["selector"] = "div.news-title a"
  field = make(map[string]interface{})
  field["name"] = "title"
  field["node"] = node
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "div.news-title a"
  node["attr"] = "href"
  field = make(map[string]interface{})
  field["name"] = "id"
  field["node"] = node
  regexReplace = make([]map[string]string, 0)
  regexReplace = append(regexReplace, map[string]string{
    "pattern": "/article/([^/]+).html",
    "value":   "$1",
  })
  field["regex_replace"] = regexReplace
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "span.news-category"
  field = make(map[string]interface{})
  field["name"] = "category"
  field["node"] = node
  fields = append(fields, field)

  node = make(map[string]interface{})
  node["selector"] = "span.news-published-time"
  field = make(map[string]interface{})
  field["name"] = "published-time"
  field["node"] = node
  fields = append(fields, field)

  html["fields"] = fields

  rules = make(map[string]interface{})
  rules["html"] = html
  extractRules["hot-list"] = rules

  useProxy := true
  timeout := 10
  r, err := h.Repository.Save(
    parentId,
    name,
    slug,
    url,
    headers,
    extractRules,
    useProxy,
    timeout,
  )
  if err != nil {
    return err
  }
  log.Println("result", r.Success, r.Message)
  return nil
}
