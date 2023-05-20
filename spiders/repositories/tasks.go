package repositories

import (
  "crypto/sha1"
  "encoding/hex"
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "net"
  "net/http"
  "time"

  "github.com/PuerkitoBio/goquery"
  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/models"
)

type TasksRepository struct {
  Db                *gorm.DB
  SourcesRepository *SourcesRepository
}

func (r *TasksRepository) Source() *SourcesRepository {
  if r.SourcesRepository == nil {
    r.SourcesRepository = &SourcesRepository{
      Db: r.Db,
    }
  }
  return r.SourcesRepository
}

func (r *TasksRepository) Find(id string) (*models.Task, error) {
  var entity *models.Task
  result := r.Db.First(&entity, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *TasksRepository) Get(id string) (*models.Task, error) {
  var entity *models.Task
  result := r.Db.Where("id = ?", id).First(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *TasksRepository) GetBySourceID(sourceID string) (*models.Task, error) {
  var entity *models.Task
  result := r.Db.Where("source_id", sourceID).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *TasksRepository) Save(
  parentId string,
  sourceId string,
  url string,
) error {
  hash := sha1.Sum([]byte(url))
  urlSha1 := hex.EncodeToString(hash[:])

  var entity *models.Task
  result := r.Db.Where("url_sha1 = ? AND url = ?", urlSha1, url).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = &models.Task{
      ID:            xid.New().String(),
      ParentID:      parentId,
      SourceID:      sourceId,
      Url:           url,
      UrlSha1:       urlSha1,
      ExtractResult: map[string]interface{}{},
    }
    r.Db.Create(&entity)
  } else {
    entity.ParentID = parentId
    entity.SourceID = sourceId
    entity.Status = 0
    r.Db.Model(&models.Task{ID: entity.ID}).Updates(entity)
  }

  return nil
}

func (r *TasksRepository) Process(task *models.Task) error {
  tr := &http.Transport{
    DisableKeepAlives: true,
  }

  source, err := r.Source().Get(task.SourceID)
  if err != nil {
    return err
  }

  if source.UseProxy {
    session := &common.ProxySession{
      Proxy: fmt.Sprintf("socks5://127.0.0.1:1088?timeout=%ds", source.Timeout),
    }
    tr.DialContext = session.DialContext
  } else {
    session := &net.Dialer{}
    tr.DialContext = session.DialContext
  }

  httpClient := &http.Client{
    Transport: tr,
    Timeout:   time.Duration(source.Timeout) * time.Second,
  }

  req, _ := http.NewRequest("GET", task.Url, nil)
  for key, val := range source.Headers {
    req.Header.Set(key, val.(string))
  }
  resp, err := httpClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return errors.New(
      fmt.Sprintf(
        "request error: status[%s] code[%d]",
        resp.Status,
        resp.StatusCode,
      ),
    )
  }

  var content string
  var doc *goquery.Document

  result := make(map[string]interface{})
  for key, value := range source.ExtractRules {
    rules := r.Source().ToExtractRules(value)
    if rules.Html != nil {
      if doc == nil {
        doc, err = goquery.NewDocumentFromReader(resp.Body)
        if err != nil {
          return err
        }
      }
      if rules.Html.List != nil {
        result[key], err = r.Source().ExtractHtmlList(doc, rules.Html)
      } else {
        result[key], err = r.Source().ExtractHtml(doc, rules.Html)
      }
    }
    if rules.Json != nil {
      if _, ok := result[key]; ok {
        content = result[key].(string)
      } else {
        if content == "" {
          body, _ := ioutil.ReadAll(resp.Body)
          content = string(body)
          if content == "" {
            return errors.New("content is empty")
          }
        }
      }
      if rules.Json.List != "" {
        result[key], err = r.Source().ExtractJsonList(content, rules.Json)
        if err != nil {
          continue
        }
      } else {
        result[key], err = r.Source().ExtractJson(content, rules.Json)
        if err != nil {
          continue
        }
      }
    }
  }

  task.Status = 1
  task.ExtractResult = r.JSONMap(result)

  r.Db.Model(&models.Task{ID: task.ID}).Updates(task)

  return nil
}

func (r *TasksRepository) JSONMap(in interface{}) datatypes.JSONMap {
  buf, _ := json.Marshal(in)

  var out datatypes.JSONMap
  json.Unmarshal(buf, &out)
  return out
}
