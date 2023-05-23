package workers

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/hibiken/asynq"
  "net/url"
  "strings"

  "github.com/go-redis/redis/v8"
  "github.com/nats-io/nats.go"
  "github.com/tidwall/gjson"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/repositories"
)

type Sources struct {
  Db         *gorm.DB
  Rdb        *redis.Client
  Ctx        context.Context
  Asynq      *asynq.Client
  Repository *repositories.SourcesRepository
}

func NewSources() *Sources {
  h := &Sources{
    Db:    common.NewDB(),
    Rdb:   common.NewRedis(),
    Ctx:   context.Background(),
    Asynq: common.NewAsynqClient(),
  }
  h.Repository = &repositories.SourcesRepository{
    Db:    h.Db,
    Asynq: h.Asynq,
  }
  return h
}

type SourcesProcessPayload struct {
  ID string
}

func (h *Sources) Subscribe(nc *nats.Conn) error {
  items := h.Repository.All([]string{"id", "url", "params"})
  for _, source := range items {
    if len(source.Params) == 0 {
      continue
    }
    if _, ok := source.Params["split"]; !ok {
      continue
    }
    for parent, path := range source.Params["split"].(map[string]interface{}) {
      h.Split(parent, path.([]interface{}), source.ID, source.Url, source.Params, nc)
    }
  }
  return nil
}

func (h *Sources) Split(
  parent string,
  path []interface{},
  sourceId string,
  sourceUrl string,
  sourceParams datatypes.JSONMap,
  nc *nats.Conn,
) error {
  nc.Subscribe(parent, func(m *nats.Msg) {
    task, err := h.Repository.Tasks().Get(string(m.Data))
    if err != nil {
      return
    }
    content, err := json.Marshal(task.ExtractResult)
    if err != nil {
      return
    }
    for _, split := range path {
      gjson.GetBytes(content, split.(string)).ForEach(func(_, s gjson.Result) bool {
        if strings.Contains(sourceUrl, "{}") {
          sourceUrl = strings.Replace(sourceUrl, "{}", fmt.Sprintf("%v", s.Value()), 1)
        }

        url, err := url.Parse(sourceUrl)
        if err != nil {
          return false
        }

        values := url.Query()
        if items, ok := sourceParams["query"].([]interface{}); ok {
          for _, item := range items {
            item := item.(map[string]interface{})
            name := item["name"].(string)
            value := item["value"].(string)
            if value == "$0" {
              value = s.Value().(string)
            }
            if value == "$1" {
              continue
            }
            values[name] = []string{value}
          }
        }
        url.RawQuery = values.Encode()

        err = h.Repository.Tasks().Save("", sourceId, url.String())
        if err != nil {
          return false
        }

        return true
      })
    }
  })
  return nil
}
