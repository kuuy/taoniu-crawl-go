package tasks

import "taoniu.local/crawls/spiders/repositories"

type SourcesTask struct {
  Repository *repositories.SourcesRepository
}

func (t *SourcesTask) Flush(slug string) error {
  source, err := t.Repository.GetBySlug(slug)
  if err != nil {
    return err
  }
  return t.Repository.Flush(source)
}
