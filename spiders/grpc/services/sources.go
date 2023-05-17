package services

import (
  "context"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  pb "taoniu.local/crawls/spiders/grpc/sources"
  repositories "taoniu.local/crawls/spiders/repositories"
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

func (srv *Sources) Save(ctx context.Context, request *pb.SaveRequest) (*pb.SaveReply, error) {
  reply := &pb.SaveReply{}

  headers := make(map[string]string)
  extractRules := make(map[string]*repositories.ExtractRules)

  for _, header := range request.Headers {
    headers[header.Name] = header.Value
  }

  for _, rules := range request.ExtractRules {
    extractRules[rules.Name] = &repositories.ExtractRules{}
    if rules.Html.Container.Selector != "" {
      extractRules[rules.Name].Html = &repositories.HtmlExtractRules{}
      extractRules[rules.Name].Html.Container = &repositories.HtmlExtractNode{
        Selector: rules.Html.Container.Selector,
        Attr:     rules.Html.Container.Attr,
        Index:    int(rules.Html.Container.Index),
      }
    }
    if rules.Html.List.Selector != "" {
      extractRules[rules.Name].Html.List = &repositories.HtmlExtractNode{
        Selector: rules.Html.List.Selector,
        Attr:     rules.Html.List.Attr,
        Index:    int(rules.Html.List.Index),
      }
    }
    for _, item := range rules.Html.Fields {
      field := &repositories.HtmlExtractField{}
      field.Name = item.Name
      field.Node = &repositories.HtmlExtractNode{
        Selector: item.Node.Selector,
        Attr:     item.Node.Attr,
        Index:    int(item.Node.Index),
      }
      field.Match = item.Match
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
      extractRules[rules.Name].Html.Fields = append(extractRules[rules.Name].Html.Fields, field)
    }
  }

  err := srv.Repository.Save(
    request.ParentId,
    request.Name,
    request.Slug,
    request.Url,
    headers,
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

func (srv *Sources) Register(s *grpc.Server) error {
  pb.RegisterSourcesServer(s, srv)
  return nil
}
