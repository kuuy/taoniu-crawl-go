package news

import (
  "context"
  "log"

  "github.com/urfave/cli/v2"

  services "taoniu.local/crawls/cryptos/grpc/services/spiders"
  repositories "taoniu.local/crawls/cryptos/repositories/spiders"
)

type CategoriesHandler struct {
  Ctx        context.Context
  Repository *repositories.SourcesRepository
}

func NewCategoriesCommand() *cli.Command {
  var h CategoriesHandler
  return &cli.Command{
    Name:  "categories",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = CategoriesHandler{
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

func (h *CategoriesHandler) save() error {
  log.Println("sources save processing...")
  parent := "aicoin-news"
  name := "资讯分类（AICOIN）"
  slug := "aicoin-news-categories"
  url := "https://www.aicoin.com/api/data/more?cat={}&last={}"
  headers := map[string]string{
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
    "Referer":    "https://www.aicoin.com/",
  }
  extractRules := make(map[string]interface{})

  var rules map[string]interface{}
  var json map[string]interface{}
  var fields []interface{}
  var field map[string]interface{}

  json = make(map[string]interface{})

  json["container"] = ""
  json["list"] = "data"

  fields = make([]interface{}, 0)
  fields = append(fields, map[string]interface{}{
    "name": "id",
    "path": "id",
  })
  fields = append(fields, map[string]interface{}{
    "name": "title",
    "path": "title",
  })
  fields = append(fields, map[string]interface{}{
    "name": "type",
    "path": "type",
  })
  fields = append(fields, map[string]interface{}{
    "name": "source_id",
    "path": "source_id",
  })
  fields = append(fields, map[string]interface{}{
    "name": "source_name",
    "path": "source_name",
  })
  fields = append(fields, map[string]interface{}{
    "name": "cover",
    "path": "cover",
  })
  fields = append(fields, map[string]interface{}{
    "name": "avatar",
    "path": "avatar",
  })
  fields = append(fields, map[string]interface{}{
    "name": "describe",
    "path": "describe",
  })
  fields = append(fields, map[string]interface{}{
    "name": "createtime",
    "path": "createtime",
  })
  field = make(map[string]interface{})
  field["name"] = "viewpoints"
  field["path"] = "viewpoints"
  field["fields"] = []map[string]interface{}{
    {
      "name": "key",
      "path": "key",
    },
    {
      "name": "coin_show",
      "path": "coin_show",
    },
    {
      "name": "viewpoint",
      "path": "viewpoint",
    },
  }
  fields = append(fields, field)

  json["fields"] = fields

  rules = make(map[string]interface{})
  rules["json"] = json
  extractRules["articles"] = rules

  source, err := h.Repository.GetBySlug(parent)
  if err != nil {
    return err
  }
  log.Println("source", source)

  useProxy := true
  timeout := 10
  r, err := h.Repository.Save(
    source.Data.Id,
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
