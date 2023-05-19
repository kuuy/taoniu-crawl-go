package spiders

import (
  "context"
  "fmt"
  "os"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"

  pb "taoniu.local/crawls/cryptos/grpc/spiders/protos/sources"
)

type Sources struct {
  Ctx           context.Context
  SourcesClient pb.SourcesClient
}

func (srv *Sources) Client() pb.SourcesClient {
  if srv.SourcesClient == nil {
    conn, err := grpc.Dial(
      fmt.Sprintf("%v:%v", os.Getenv("SPIDERS_GRPC_HOST"), os.Getenv("SPIDERS_GRPC_PORT")),
      grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
      panic(err.Error())
    }
    srv.SourcesClient = pb.NewSourcesClient(conn)
  }
  return srv.SourcesClient
}

func (srv *Sources) Get(id string) (*pb.GetReply, error) {
  request := &pb.GetRequest{
    Id: id,
  }

  r, err := srv.Client().Get(srv.Ctx, request)
  if err != nil {
    return nil, err
  }

  return r, nil
}

func (srv *Sources) GetBySlug(slug string) (*pb.GetBySlugReply, error) {
  request := &pb.GetBySlugRequest{
    Slug: slug,
  }

  r, err := srv.Client().GetBySlug(srv.Ctx, request)
  if err != nil {
    return nil, err
  }

  return r, nil
}

func (srv *Sources) Save(
  parentId string,
  name string,
  slug string,
  url string,
  headers map[string]string,
  extractRules map[string]interface{},
  useProxy bool,
  timeout int,
) (*pb.SaveReply, error) {
  request := &pb.SaveRequest{
    ParentId: parentId,
    Name:     name,
    Slug:     slug,
    Url:      url,
    UseProxy: useProxy,
    Timeout:  uint32(timeout),
  }

  for name, value := range headers {
    request.Headers = append(request.Headers, &pb.HttpHeader{
      Name:  name,
      Value: value,
    })
  }

  for name, data := range extractRules {
    rules := &pb.ExtractRules{
      Name: name,
    }

    data := data.(map[string]interface{})

    if html, ok := data["html"]; ok {
      rules.Html = &pb.HtmlExtractRules{}
      html := html.(map[string]interface{})
      if container, ok := html["container"]; ok {
        container := container.(map[string]interface{})
        rules.Html.Container = &pb.HtmlExtractNode{}
        if value, ok := container["selector"]; ok {
          rules.Html.Container.Selector = value.(string)
        }
        if value, ok := container["attr"]; ok {
          rules.Html.Container.Attr = value.(string)
        }
        if value, ok := container["index"]; ok {
          rules.Html.Container.Index = value.(uint32)
        }
      }
      if list, ok := html["list"]; ok {
        list := list.(map[string]interface{})
        rules.Html.List = &pb.HtmlExtractNode{}
        if value, ok := list["selector"]; ok {
          rules.Html.List.Selector = value.(string)
        }
        if value, ok := list["attr"]; ok {
          rules.Html.List.Attr = value.(string)
        }
        if value, ok := list["index"]; ok {
          rules.Html.List.Index = value.(uint32)
        }
      }
      if fields, ok := html["fields"]; ok {
        fields := fields.([]interface{})
        for _, field := range fields {
          field := field.(map[string]interface{})
          message := &pb.HtmlExtractField{}
          if value, ok := field["name"]; ok {
            message.Name = value.(string)
          }
          if value, ok := field["node"]; ok {
            value := value.(map[string]interface{})
            message.Node = &pb.HtmlExtractNode{}
            if value, ok := value["selector"]; ok {
              message.Node.Selector = value.(string)
            }
            if value, ok := value["attr"]; ok {
              message.Node.Attr = value.(string)
            }
            if value, ok := value["index"]; ok {
              message.Node.Index = value.(uint32)
            }
          }
          if value, ok := field["match"]; ok {
            message.Match = value.(string)
          }
          if values, ok := field["regex_replace"]; ok {
            values := values.([]map[string]string)
            for _, value := range values {
              replace := &pb.RegexReplace{}
              if value, ok := value["pattern"]; ok {
                replace.Pattern = value
              }
              if value, ok := value["value"]; ok {
                replace.Value = value
              }
              message.RegexReplace = append(message.RegexReplace, replace)
            }
          }
          if values, ok := field["text_replace"]; ok {
            values := values.([]map[string]string)
            for _, value := range values {
              replace := &pb.TextReplace{}
              if value, ok := value["text"]; ok {
                replace.Text = value
              }
              if value, ok := value["value"]; ok {
                replace.Value = value
              }
              message.TextReplace = append(message.TextReplace, replace)
            }
          }
          rules.Html.Fields = append(rules.Html.Fields, message)
        }
      }
    }

    if json, ok := data["json"]; ok {
      rules.Json = &pb.JsonExtractRules{}
      json := json.(map[string]interface{})
      if container, ok := json["container"]; ok {
        rules.Json.Container = container.(string)
      }
      if list, ok := json["list"]; ok {
        rules.Json.List = list.(string)
      }
      if fields, ok := json["fields"]; ok {
        fields := fields.([]interface{})
        for _, field := range fields {
          field := field.(map[string]interface{})
          message := &pb.JsonExtractField{}
          if value, ok := field["name"]; ok {
            message.Name = value.(string)
          }
          if value, ok := field["path"]; ok {
            message.Path = value.(string)
          }
          if value, ok := field["match"]; ok {
            message.Match = value.(string)
          }
          if values, ok := field["regex_replace"]; ok {
            values := values.([]map[string]string)
            for _, value := range values {
              replace := &pb.RegexReplace{}
              if value, ok := value["pattern"]; ok {
                replace.Pattern = value
              }
              if value, ok := value["value"]; ok {
                replace.Value = value
              }
              message.RegexReplace = append(message.RegexReplace, replace)
            }
          }
          if values, ok := field["text_replace"]; ok {
            values := values.([]map[string]string)
            for _, value := range values {
              replace := &pb.TextReplace{}
              if value, ok := value["text"]; ok {
                replace.Text = value
              }
              if value, ok := value["value"]; ok {
                replace.Value = value
              }
              message.TextReplace = append(message.TextReplace, replace)
            }
          }
          rules.Json.Fields = append(rules.Json.Fields, message)
        }
      }
    }

    request.ExtractRules = append(request.ExtractRules, rules)
  }

  r, err := srv.Client().Save(srv.Ctx, request)
  if err != nil {
    return nil, err
  }

  return r, nil
}
