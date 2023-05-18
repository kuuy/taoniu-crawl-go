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
  "regexp"
  "strings"
  "time"

  "github.com/PuerkitoBio/goquery"
  "github.com/rs/xid"
  "github.com/tidwall/gjson"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/models"
)

type SourcesRepository struct {
  Db *gorm.DB
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
  Name         string           `json:"name"`
  Node         *HtmlExtractNode `json:"node"`
  Match        string           `json:"match"`
  RegexReplace []*RegexReplace  `json:"regex_replace"`
  TextReplace  []*TextReplace   `json:"text_replace"`
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
  Name         string          `json:"name"`
  Path         string          `json:"path"`
  Match        string          `json:"match"`
  RegexReplace []*RegexReplace `json:"regex_replace"`
  TextReplace  []*TextReplace  `json:"text_replace"`
}

func (r *SourcesRepository) Find(id string) (*models.Source, error) {
  var entity *models.Source
  result := r.Db.First(&entity, "id", id)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, result.Error
  }
  return entity, nil
}

func (r *SourcesRepository) Get(slug string) (*models.Source, error) {
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
  useProxy bool,
  timeout int,
  extractRules map[string]*ExtractRules,
) error {
  hash := sha1.Sum([]byte(url))

  var entity *models.Source
  result := r.Db.Where("slug", slug).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = &models.Source{
      ID:            xid.New().String(),
      ParentID:      parentId,
      Name:          name,
      Slug:          slug,
      Url:           url,
      UrlSha1:       hex.EncodeToString(hash[:]),
      Headers:       r.JSONMap(headers),
      UseProxy:      useProxy,
      Timeout:       timeout,
      ExtractRules:  r.JSONMap(extractRules),
      ExtractResult: r.JSONMap(make(map[string]interface{})),
    }
    r.Db.Create(&entity)
  } else {
    entity.ParentID = parentId
    entity.Name = name
    entity.Url = url
    entity.UrlSha1 = hex.EncodeToString(hash[:])
    entity.Headers = r.JSONMap(headers)
    entity.UseProxy = useProxy
    entity.Timeout = timeout
    entity.ExtractRules = r.JSONMap(extractRules)
    r.Db.Model(&models.Source{ID: entity.ID}).Updates(entity)
  }

  return nil
}

func (r *SourcesRepository) Flush(source *models.Source) error {
  tr := &http.Transport{
    DisableKeepAlives: true,
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

  req, _ := http.NewRequest("GET", source.Url, nil)
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
    rules := r.ToExtractRules(value)
    if rules.Html != nil {
      if doc == nil {
        doc, err = goquery.NewDocumentFromReader(resp.Body)
        if err != nil {
          return err
        }
      }
      if rules.Html.List != nil {
        result[key], err = r.ExtractHtmlList(doc, rules.Html)
      } else {
        result[key], err = r.ExtractHtml(doc, rules.Html)
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
        result[key], err = r.ExtractJsonList(content, rules.Json)
        if err != nil {
          continue
        }
      } else {
        result[key], err = r.ExtractJson(content, rules.Json)
        if err != nil {
          continue
        }
      }
    }
  }

  source.ExtractResult = r.JSONMap(result)

  r.Db.Model(&models.Source{ID: source.ID}).Updates(source)

  return nil
}

func (r *SourcesRepository) ExtractHtml(doc *goquery.Document, rules *HtmlExtractRules) (map[string]string, error) {
  var container = doc.Find(rules.Container.Selector).Eq(rules.Container.Index)
  if container.Nodes == nil {
    return nil, errors.New("container not exists")
  }

  var data = make(map[string]string)
  for _, field := range rules.Fields {
    if field.Node.Selector != "" {
      selection := container.Find(field.Node.Selector).Eq(field.Node.Index)
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
        attr, exists := container.Attr(field.Node.Attr)
        if !exists {
          continue
        }
        data[field.Name] = attr
      } else {
        data[field.Name] = container.Text()
      }
    }

    for _, replace := range field.RegexReplace {
      m := regexp.MustCompile(replace.Pattern)
      data[field.Name] = m.ReplaceAllString(data[field.Name], replace.Value)
    }
    for _, replace := range field.TextReplace {
      data[field.Name] = strings.ReplaceAll(data[field.Name], replace.Text, replace.Value)
    }

    if field.Match != "" && field.Match != data[field.Name] {
      return nil, errors.New("field not match")
    }
  }

  return data, nil
}

func (r *SourcesRepository) ExtractHtmlList(doc *goquery.Document, rules *HtmlExtractRules) ([]map[string]string, error) {
  var container = doc.Find(rules.Container.Selector).Eq(rules.Container.Index)
  if container.Nodes == nil {
    return nil, errors.New("container not exists")
  }

  var result []map[string]string
  container.Find(rules.List.Selector).Each(func(i int, s *goquery.Selection) {
    var data = make(map[string]string)
    for _, field := range rules.Fields {
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

      for _, replace := range field.RegexReplace {
        m := regexp.MustCompile(replace.Pattern)
        data[field.Name] = m.ReplaceAllString(data[field.Name], replace.Value)
      }
      for _, replace := range field.TextReplace {
        data[field.Name] = strings.ReplaceAll(data[field.Name], replace.Text, replace.Value)
      }

      if field.Match != "" && field.Match != data[field.Name] {
        break
      }
    }
    result = append(result, data)
  })

  return result, nil
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

func (r *SourcesRepository) ExtractJsonList(content string, rules *JsonExtractRules) ([]map[string]interface{}, error) {
  var container = gjson.Get(content, rules.Container)
  if container.Raw == "" {
    return nil, errors.New("container not exists")
  }

  var result []map[string]interface{}
  container.Get(rules.List).ForEach(func(_, s gjson.Result) bool {
    var data = make(map[string]interface{})
    for _, field := range rules.Fields {
      selection := s.Get(field.Path)
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
          return false
        }
      }
    }
    result = append(result, data)

    return true
  })

  return result, nil
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
