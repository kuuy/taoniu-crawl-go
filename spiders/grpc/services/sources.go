package services

import (
  "context"
  "google.golang.org/grpc"
  "google.golang.org/protobuf/types/known/timestamppb"
  "gorm.io/gorm"
  pb "taoniu.local/crawls/spiders/grpc/sources"
  "taoniu.local/crawls/spiders/repositories"
)

type Sources struct {
  pb.UnimplementedSourcesServer
  Repository *repositories.SourcesRepository
}

func NewSources(db *gorm.DB) *Sources {
  return &Sources{
    Repository: &repositories.SourcesRepository{
      Db: db,
    },
  }
}

func (srv *Sources) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetReply, error) {
  reply := &pb.GetReply{}
  source, err := srv.Repository.Get(request.Id)
  if err != nil {
    return nil, err
  }

  reply.Data = &pb.SourceInfo{
    Id:        source.ID,
    Name:      source.Name,
    Slug:      source.Slug,
    Url:       source.Url,
    UseProxy:  source.UseProxy,
    Timeout:   int32(source.Timeout),
    Status:    int32(source.Status),
    CreatedAt: timestamppb.New(source.CreatedAt),
    UpdatedAt: timestamppb.New(source.UpdatedAt),
  }

  for name, value := range source.Headers {
    reply.Data.Headers = append(reply.Data.Headers, &pb.HttpHeader{
      Name:  name,
      Value: value.(string),
    })
  }

  for name, data := range source.ExtractRules {
    rules := srv.Repository.ToExtractRules(data)
    reply.Data.ExtractRules = append(reply.Data.ExtractRules, srv.ToExtractRules(name, rules))
  }

  return reply, nil
}

func (srv *Sources) GetBySlug(ctx context.Context, request *pb.GetBySlugRequest) (*pb.GetBySlugReply, error) {
  reply := &pb.GetBySlugReply{}
  source, err := srv.Repository.GetBySlug(request.Slug)
  if err != nil {
    return nil, err
  }

  reply.Data = &pb.SourceInfo{
    Id:        source.ID,
    Name:      source.Name,
    Slug:      source.Slug,
    Url:       source.Url,
    UseProxy:  source.UseProxy,
    Timeout:   int32(source.Timeout),
    Status:    int32(source.Status),
    CreatedAt: timestamppb.New(source.CreatedAt),
    UpdatedAt: timestamppb.New(source.UpdatedAt),
  }

  for name, value := range source.Headers {
    reply.Data.Headers = append(reply.Data.Headers, &pb.HttpHeader{
      Name:  name,
      Value: value.(string),
    })
  }

  for name, data := range source.ExtractRules {
    rules := srv.Repository.ToExtractRules(data)
    reply.Data.ExtractRules = append(reply.Data.ExtractRules, srv.ToExtractRules(name, rules))
  }

  return reply, nil
}

func (srv *Sources) Save(ctx context.Context, request *pb.SaveRequest) (*pb.SaveReply, error) {
  reply := &pb.SaveReply{}

  headers := map[string]string{}
  for _, header := range request.Headers {
    headers[header.Name] = header.Value
  }

  params := map[string]interface{}{}

  var split []string
  for _, value := range request.Params.Split {
    split = append(split, value)
  }
  if len(split) > 0 {
    params["split"] = split
  }

  if request.Params.Scroll != "" {
    params["scroll"] = request.Params.Scroll
  }

  var query []map[string]string
  for _, item := range request.Params.Query {
    query = append(query, map[string]string{
      "name":    item.Name,
      "value":   item.Value,
      "default": item.Default,
    })
  }
  if len(query) > 0 {
    params["query"] = query
  }

  extractRules := map[string]*repositories.ExtractRules{}
  for _, rules := range request.ExtractRules {
    extractRules[rules.Name] = srv.MapExtractRules(rules)
  }

  err := srv.Repository.Save(
    request.ParentId,
    request.Name,
    request.Slug,
    request.Url,
    headers,
    params,
    request.UseProxy,
    int(request.Timeout),
    extractRules,
  )
  if err != nil {
    reply.Message = err.Error()
  } else {
    reply.Success = true
  }
  return reply, nil
}

func (srv *Sources) MapExtractRules(data *pb.ExtractRules) *repositories.ExtractRules {
  rules := &repositories.ExtractRules{}
  if data.Html != nil {
    rules.Html = srv.MapHtmlExtractRules(data.Html)
  }
  if data.Json != nil {
    rules.Json = srv.MapJsonExtractRules(data.Json)
  }
  return rules
}

func (srv *Sources) MapHtmlExtractRules(data *pb.HtmlExtractRules) *repositories.HtmlExtractRules {
  rules := &repositories.HtmlExtractRules{}
  rules.Container = &repositories.HtmlExtractNode{
    Selector: data.Container.Selector,
    Attr:     data.Container.Attr,
    Index:    int(data.Container.Index),
  }
  if data.List != nil {
    rules.List = &repositories.HtmlExtractNode{
      Selector: data.List.Selector,
      Attr:     data.List.Attr,
      Index:    int(data.List.Index),
    }
  }
  rules.Fields = srv.MapHtmlExtractField(data.Fields)

  return rules
}

func (srv *Sources) MapHtmlExtractField(items []*pb.HtmlExtractField) []*repositories.HtmlExtractField {
  var fields []*repositories.HtmlExtractField
  for _, item := range items {
    field := &repositories.HtmlExtractField{
      Name:  item.Name,
      Match: item.Match,
    }
    field.Node = &repositories.HtmlExtractNode{
      Selector: item.Node.Selector,
      Attr:     item.Node.Attr,
      Index:    int(item.Node.Index),
    }

    if len(item.Fields) > 0 {
      field.Fields = srv.MapHtmlExtractField(item.Fields)
    }

    for _, replace := range item.RegexReplace {
      field.RegexReplace = append(field.RegexReplace, &repositories.RegexReplace{
        Pattern: replace.Pattern,
        Value:   replace.Value,
      })
    }

    for _, replace := range item.TextReplace {
      field.TextReplace = append(field.TextReplace, &repositories.TextReplace{
        Text:  replace.Text,
        Value: replace.Value,
      })
    }

    fields = append(fields, field)
  }

  return fields
}

func (srv *Sources) MapJsonExtractRules(data *pb.JsonExtractRules) *repositories.JsonExtractRules {
  rules := &repositories.JsonExtractRules{
    Container: data.Container,
    List:      data.List,
  }
  rules.Fields = srv.MapJsonExtractField(data.Fields)
  return rules
}

func (srv *Sources) MapJsonExtractField(items []*pb.JsonExtractField) []*repositories.JsonExtractField {
  var fields []*repositories.JsonExtractField
  for _, item := range items {
    field := &repositories.JsonExtractField{
      Name:  item.Name,
      Path:  item.Path,
      Match: item.Match,
    }

    if len(item.Fields) > 0 {
      field.Fields = srv.MapJsonExtractField(item.Fields)
    }

    for _, replace := range item.RegexReplace {
      field.RegexReplace = append(field.RegexReplace, &repositories.RegexReplace{
        Pattern: replace.Pattern,
        Value:   replace.Value,
      })
    }

    for _, replace := range item.TextReplace {
      field.TextReplace = append(field.TextReplace, &repositories.TextReplace{
        Text:  replace.Text,
        Value: replace.Value,
      })
    }

    fields = append(fields, field)
  }

  return fields
}

func (srv *Sources) ToExtractRules(name string, data *repositories.ExtractRules) *pb.ExtractRules {
  rules := &pb.ExtractRules{
    Name: name,
  }
  if data.Html != nil {
    rules.Html = srv.ToHtmlExtractRules(data.Html)
  }
  if data.Json != nil {
    rules.Json = srv.ToJsonExtractRules(data.Json)
  }
  return rules
}

func (srv *Sources) ToHtmlExtractRules(data *repositories.HtmlExtractRules) *pb.HtmlExtractRules {
  rules := &pb.HtmlExtractRules{}
  rules.Container = &pb.HtmlExtractNode{
    Selector: data.Container.Selector,
    Attr:     data.Container.Attr,
    Index:    uint32(data.Container.Index),
  }
  if data.List != nil {
    rules.List = &pb.HtmlExtractNode{
      Selector: data.Container.Selector,
      Attr:     data.Container.Attr,
      Index:    uint32(data.Container.Index),
    }
  }
  rules.Fields = srv.ToHtmlExtractField(data.Fields)

  return rules
}

func (srv *Sources) ToHtmlExtractField(items []*repositories.HtmlExtractField) []*pb.HtmlExtractField {
  var fields []*pb.HtmlExtractField
  for _, item := range items {
    field := &pb.HtmlExtractField{
      Name:  item.Name,
      Match: item.Match,
    }
    field.Node = &pb.HtmlExtractNode{
      Selector: item.Node.Selector,
      Attr:     item.Node.Attr,
      Index:    uint32(item.Node.Index),
    }

    if len(item.Fields) > 0 {
      field.Fields = srv.ToHtmlExtractField(item.Fields)
    }

    for _, replace := range item.RegexReplace {
      field.RegexReplace = append(field.RegexReplace, &pb.RegexReplace{
        Pattern: replace.Pattern,
        Value:   replace.Value,
      })
    }

    for _, replace := range item.TextReplace {
      field.TextReplace = append(field.TextReplace, &pb.TextReplace{
        Text:  replace.Text,
        Value: replace.Value,
      })
    }

    fields = append(fields, field)
  }
  return fields
}

func (srv *Sources) ToJsonExtractRules(data *repositories.JsonExtractRules) *pb.JsonExtractRules {
  rules := &pb.JsonExtractRules{
    Container: data.Container,
    List:      data.List,
  }
  rules.Fields = srv.ToJsonExtractField(data.Fields)
  return rules
}

func (srv *Sources) ToJsonExtractField(items []*repositories.JsonExtractField) []*pb.JsonExtractField {
  var fields []*pb.JsonExtractField
  for _, item := range items {
    field := &pb.JsonExtractField{
      Name:  item.Name,
      Path:  item.Path,
      Match: item.Match,
    }

    if len(item.Fields) > 0 {
      field.Fields = srv.ToJsonExtractField(item.Fields)
    }

    for _, replace := range item.RegexReplace {
      field.RegexReplace = append(field.RegexReplace, &pb.RegexReplace{
        Pattern: replace.Pattern,
        Value:   replace.Value,
      })
    }

    for _, replace := range item.TextReplace {
      field.TextReplace = append(field.TextReplace, &pb.TextReplace{
        Text:  replace.Text,
        Value: replace.Value,
      })
    }

    fields = append(fields, field)
  }
  return fields
}

func (srv *Sources) Register(s *grpc.Server) error {
  pb.RegisterSourcesServer(s, srv)
  return nil
}
