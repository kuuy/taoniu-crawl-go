package spiders

import (
  services "taoniu.local/crawls/cryptos/grpc/services/spiders"
  pb "taoniu.local/crawls/cryptos/grpc/spiders/protos/sources"
)

type SourcesRepository struct {
  Service *services.Sources
}

func (r *SourcesRepository) Save(
  parentId string,
  name string,
  slug string,
  url string,
  headers map[string]string,
  extractRules map[string]interface{},
  useProxy bool,
  timeout int,
) (*pb.SaveReply, error) {
  return r.Service.Save(parentId, name, slug, url, headers, extractRules, useProxy, timeout)
}

func (r *SourcesRepository) Get(id string) (*pb.GetReply, error) {
  return r.Service.Get(id)
}

func (r *SourcesRepository) GetBySlug(slug string) (*pb.GetBySlugReply, error) {
  return r.Service.GetBySlug(slug)
}
