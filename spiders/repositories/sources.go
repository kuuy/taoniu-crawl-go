package repositories

import (
  "crypto/sha1"
  "encoding/hex"
  "encoding/json"
  "errors"
  "fmt"
  "net"
  "net/http"
  "regexp"
  "time"

  "github.com/PuerkitoBio/goquery"
  "github.com/rs/xid"
  "gorm.io/datatypes"
  "gorm.io/gorm"

  "taoniu.local/crawls/spiders/common"
  "taoniu.local/crawls/spiders/models"
)

type SourcesRepository struct {
  Db *gorm.DB
}

type HtmlExtractField struct {
  Name         string           `json:"name"`
  Node         *HtmlExtractNode `json:"node"`
  RegexReplace []*RegexReplace  `json:"replace"`
}

type HtmlExtractNode struct {
  Selector string `json:"selector"`
  Attr     string `json:"attr"`
  Index    int    `json:"index"`
}

type RegexReplace struct {
  Pattern string `json:"pattern"`
  Value   string `json:"replace"`
}

type ExtractRules struct {
  Container *HtmlExtractNode    `json:"container"`
  List      *HtmlExtractNode    `json:"list"`
  Json      []*JsonExtract      `json:"json"`
  Fields    []*HtmlExtractField `json:"fields"`
}

type JsonExtract struct {
  Node  *HtmlExtractNode  `json:"node"`
  Rules *JsonExtractRules `json:"rules"`
}

type JsonExtractField struct {
  Name  string `json:"name"`
  Path  string `json:"path"`
  Match string `json:"match"`
}

type JsonExtractRules struct {
  Container string              `json:"container"`
  List      string              `json:"list"`
  Fields    []*JsonExtractField `json:"fields"`
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

func (r *SourcesRepository) Add(
    parentId string,
    name string,
    slug string,
    source *models.Source,
) error {
  hash := sha1.Sum([]byte(source.Url))

  var entity *models.Source
  result := r.Db.Where("slug", slug).Take(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = &models.Source{
      ID:            xid.New().String(),
      ParentID:      parentId,
      Name:          name,
      Slug:          slug,
      Url:           source.Url,
      UrlSha1:       hex.EncodeToString(hash[:]),
      Headers:       r.JSONMap(source.Headers),
      UseProxy:      source.UseProxy,
      Timeout:       source.Timeout,
      ExtractRules:  r.JSONMap(source.ExtractRules),
      ExtractResult: r.JSONMap(make(map[string]interface{})),
    }
    r.Db.Create(&entity)
  } else {
    entity.ParentID = parentId
    entity.Name = name
    entity.Url = source.Url
    entity.UrlSha1 = hex.EncodeToString(hash[:])
    entity.Headers = r.JSONMap(source.Headers)
    entity.UseProxy = source.UseProxy
    entity.Timeout = source.Timeout
    entity.ExtractRules = r.JSONMap(source.ExtractRules)
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

  doc, err := goquery.NewDocumentFromReader(resp.Body)
  if err != nil {
    return err
  }

  result := make(map[string]interface{})
  for key, rules := range source.ExtractRules {
    result[key], err = r.Extract(doc, r.ToExtractRules(rules))
    if err != nil {
      continue
    }
  }

  source.ExtractResult = r.JSONMap(result)

  r.Db.Model(&models.Source{ID: source.ID}).Updates(source)

  return nil
}

func (r *SourcesRepository) Extract(doc *goquery.Document, rules *ExtractRules) ([]map[string]string, error) {
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
        if field.Node.Attr != "" {
          data[field.Name], _ = selection.Attr(field.Node.Attr)
        } else {
          data[field.Name] = selection.Text()
        }
      } else {
        if field.Node.Attr != "" {
          data[field.Name], _ = s.Attr(field.Node.Attr)
        } else {
          data[field.Name] = s.Text()
        }
      }
      for _, replace := range field.RegexReplace {
        m := regexp.MustCompile(replace.Pattern)
        data[field.Name] = m.ReplaceAllString(data[field.Name], replace.Value)
      }
    }
    result = append(result, data)
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
