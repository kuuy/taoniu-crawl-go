package repositories

import (
  "encoding/json"
  "errors"
  "net/url"
  "regexp"
  "strings"

  "github.com/PuerkitoBio/goquery"
  "github.com/hibiken/asynq"
  "github.com/rs/xid"
  "github.com/tidwall/gjson"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "taoniu.local/crawls/spiders/models"
  "taoniu.local/crawls/spiders/queue/asynq/jobs"
)

type SourcesRepository struct {
  Db              *gorm.DB
  Asynq           *asynq.Client
  TasksRepository *TasksRepository
}

type ExtractRules struct {
  Html *HtmlExtractRules `json:"html"`
  Json *JsonExtractRules `json:"json"`
}

type HtmlExtractRules struct {
  Container *HtmlExtractNode    `json:"container"`
  List      *HtmlExtractNode    `json:"list"`
  Fields    []*HtmlExtractField `json:"fields"`
}

type HtmlExtractNode struct {
  Selector string `json:"selector"`
  Attr     string `json:"attr"`
  Index    int    `json:"index"`
}

type HtmlExtractField struct {
  Name         string              `json:"name"`
  Node         *HtmlExtractNode    `json:"node"`
  Match        string              `json:"match"`
  RegexReplace []*RegexReplace     `json:"regex_replace"`
  TextReplace  []*TextReplace      `json:"text_replace"`
  Fields       []*HtmlExtractField `json:"fields"`
}

type RegexReplace struct {
  Pattern string `json:"pattern"`
  Value   string `json:"value"`
}

type TextReplace struct {
  Text  string `json:"text"`
  Value string `json:"value"`
}

type JsonExtractRules struct {
  Container string              `json:"node"`
  List      string              `json:"list"`
  Fields    []*JsonExtractField `json:"fields"`
}

type JsonExtractField struct {
  Name         string              `json:"name"`
  Path         string              `json:"path"`
  Match        string              `json:"match"`
  RegexReplace []*RegexReplace     `json:"regex_replace"`
  TextReplace  []*TextReplace      `json:"text_replace"`
  Fields       []*JsonExtractField `json:"fields"`
}

func (r *SourcesRepository) Tasks() *TasksRepository {
  if r.TasksRepository == nil {
    r.TasksRepository = &TasksRepository{
      Db:    r.Db,
      Asynq: r.Asynq,
      Job:   &jobs.Tasks{},
    }
  }
  return r.TasksRepository
}

func (r *SourcesRepository) Find(id string) (*models.Source, error) {
  var entity *models.Source
  result := r.Db.First(&entity, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) All(fields []string) []*models.Source {
  var sources []*models.Source
  r.Db.Select(fields).Find(&sources)
  return sources
}

func (r *SourcesRepository) Get(id string) (*models.Source, error) {
  var entity *models.Source
  result := r.Db.Where("id = ?", id).First(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) GetBySlug(slug string) (*models.Source, error) {
  var entity *models.Source
  result := r.Db.Where("slug", slug).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) Save(
  parentId string,
  name string,
  slug string,
  url string,
  headers map[string]string,
  params map[string]interface{},
  useProxy bool,
  timeout int,
  extractRules map[string]*ExtractRules,
) error {
  var entity *models.Source
  result := r.Db.Where("slug", slug).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = &models.Source{
      ID:           xid.New().String(),
      ParentID:     parentId,
      Name:         name,
      Slug:         slug,
      Url:          url,
      Headers:      r.JSONMap(headers),
      Params:       r.JSONMap(params),
      UseProxy:     useProxy,
      Timeout:      timeout,
      ExtractRules: r.JSONMap(extractRules),
    }
    r.Db.Create(&entity)
  } else {
    entity.ParentID = parentId
    entity.Name = name
    entity.Url = url
    entity.Headers = r.JSONMap(headers)
    entity.Params = r.JSONMap(params)
    entity.UseProxy = useProxy
    entity.Timeout = timeout
    entity.ExtractRules = r.JSONMap(extractRules)
    r.Db.Model(&models.Source{ID: entity.ID}).Updates(entity)
  }

  return nil
}

func (r *SourcesRepository) Flush(source *models.Source) error {
  if len(source.Params) == 0 {
    return r.Tasks().Save("", source.ID, source.Url)
  }

  task, err := r.Tasks().GetBySourceID(source.ParentID)
  if err != nil {
    return err
  }
  content, err := json.Marshal(task.ExtractResult)
  if err != nil {
    return err
  }
  if items, ok := source.Params["split"].([]interface{}); ok {
    for _, split := range items {
      gjson.GetBytes(content, split.(string)).ForEach(func(_, s gjson.Result) bool {
        url, err := url.Parse(source.Url)
        if err != nil {
          return false
        }
        values := url.Query()
        if items, ok := source.Params["query"].([]interface{}); ok {
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
        r.Tasks().Save("", source.ID, url.String())
        return true
      })
    }
  }
  //source.ExtractResult = r.JSONMap(result)

  //r.Db.Model(&models.Source{ID: source.ID}).Updates(source)

  return nil
}

func (r *SourcesRepository) ExtractHtml(doc *goquery.Document, rules *HtmlExtractRules) (data map[string]interface{}, err error) {
  var container = doc.Find(rules.Container.Selector).Eq(rules.Container.Index)
  if container.Nodes == nil {
    err = errors.New("container not exists")
    return
  }

  data, err = r.ExtractHtmlFields(container, rules.Fields)
  if err != nil {
    return
  }

  return
}

func (r *SourcesRepository) ExtractHtmlList(doc *goquery.Document, rules *HtmlExtractRules) (result []map[string]interface{}, err error) {
  var container = doc.Find(rules.Container.Selector).Eq(rules.Container.Index)
  if container.Nodes == nil {
    err = errors.New("container not exists")
    return
  }

  container.Find(rules.List.Selector).Each(func(i int, s *goquery.Selection) {
    data, err := r.ExtractHtmlFields(s, rules.Fields)
    if err != nil {
      return
    }
    result = append(result, data)
  })

  return
}

func (r *SourcesRepository) ExtractHtmlFields(s *goquery.Selection, fields []*HtmlExtractField) (data map[string]interface{}, err error) {
  data = make(map[string]interface{})
  for _, field := range fields {
    if field.Node.Selector != "" {
      selection := s.Find(field.Node.Selector).Eq(field.Node.Index)
      if selection.Nodes == nil {
        continue
      }
      if field.Node.Attr != "" {
        data[field.Name], _ = selection.Attr(field.Node.Attr)
      } else {
        data[field.Name] = selection.Text()
      }
    } else {
      if field.Node.Attr != "" {
        attr, exists := s.Attr(field.Node.Attr)
        if !exists {
          continue
        }
        data[field.Name] = attr
      } else {
        data[field.Name] = s.Text()
      }
    }

    if len(field.Fields) > 0 {
      result, err := r.ExtractHtmlFields(s, field.Fields)
      if err != nil {
        continue
      }
      data[field.Name] = result
    }

    for _, replace := range field.RegexReplace {
      m := regexp.MustCompile(replace.Pattern)
      data[field.Name] = m.ReplaceAllString(data[field.Name].(string), replace.Value)
    }
    for _, replace := range field.TextReplace {
      data[field.Name] = strings.ReplaceAll(data[field.Name].(string), replace.Text, replace.Value)
    }

    if field.Match != "" && field.Match != data[field.Name] {
      err = errors.New("field not match")
      return
    }
  }

  return
}

func (r *SourcesRepository) ExtractJson(content string, rules *JsonExtractRules) (map[string]interface{}, error) {
  var container = gjson.Get(content, rules.Container)
  if container.Raw == "" {
    return nil, errors.New("container not exists")
  }

  var data = make(map[string]interface{})
  for _, field := range rules.Fields {
    selection := container.Get(field.Path)
    if selection.Raw == "" {
      continue
    }
    data[field.Name] = selection.Value()

    if value, ok := data[field.Name].(string); ok {
      for _, replace := range field.RegexReplace {
        m := regexp.MustCompile(replace.Pattern)
        data[field.Name] = m.ReplaceAllString(value, replace.Value)
      }
      for _, replace := range field.TextReplace {
        data[field.Name] = strings.ReplaceAll(value, replace.Text, replace.Value)
      }

      if field.Match != "" && field.Match != data[field.Name] {
        return nil, errors.New("field not match")
      }
    }
  }

  return data, nil
}

func (r *SourcesRepository) ExtractJsonList(content string, rules *JsonExtractRules) (result []map[string]interface{}, err error) {
  if rules.Container != "" {
    var container = gjson.Get(content, rules.Container)
    if container.Raw == "" {
      err = errors.New("container not exists")
      return
    }

    container.Get(rules.List).ForEach(func(_, s gjson.Result) bool {
      data, err := r.ExtractJsonFields(&s, rules.Fields)
      if err != nil {
        return false
      }
      result = append(result, data)
      return true
    })
  } else {
    gjson.Get(content, rules.List).ForEach(func(_, s gjson.Result) bool {
      data, err := r.ExtractJsonFields(&s, rules.Fields)
      if err != nil {
        return false
      }
      result = append(result, data)
      return true
    })
  }

  return
}

func (r *SourcesRepository) ExtractJsonFields(s *gjson.Result, fields []*JsonExtractField) (data map[string]interface{}, err error) {
  data = make(map[string]interface{})
  for _, field := range fields {
    selection := s.Get(field.Path)
    if selection.Raw == "" {
      continue
    }
    data[field.Name] = selection.Value()

    if len(field.Fields) > 0 {
      if selection.IsArray() {
        var result []map[string]interface{}
        selection.ForEach(func(_, s gjson.Result) bool {
          data, err := r.ExtractJsonFields(&s, field.Fields)
          if err != nil {
            return false
          }
          result = append(result, data)
          return true
        })
        data[field.Name] = result
      } else {
        result, err := r.ExtractJsonFields(&selection, field.Fields)
        if err != nil {
          continue
        }
        data[field.Name] = result
      }
    }

    for _, replace := range field.RegexReplace {
      m := regexp.MustCompile(replace.Pattern)
      data[field.Name] = m.ReplaceAllString(data[field.Name].(string), replace.Value)
    }
    for _, replace := range field.TextReplace {
      data[field.Name] = strings.ReplaceAll(data[field.Name].(string), replace.Text, replace.Value)
    }

    if field.Match != "" && field.Match != data[field.Name] {
      err = errors.New("field not match")
      return
    }
  }

  return
}

func (r *SourcesRepository) ToExtractRules(in interface{}) *ExtractRules {
  buf, _ := json.Marshal(in)

  var out *ExtractRules
  json.Unmarshal(buf, &out)
  return out
}

func (r *SourcesRepository) JSONMap(in interface{}) datatypes.JSONMap {
  buf, _ := json.Marshal(in)

  var out datatypes.JSONMap
  json.Unmarshal(buf, &out)
  return out
}
