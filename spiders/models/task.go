package models

import (
  "gorm.io/datatypes"
  "time"
)

type Task struct {
  ID            string            `gorm:"size:20;primaryKey"`
  ParentID      string            `gorm:"size:20;not null;index"`
  SourceID      string            `gorm:"size:20;not null;index"`
  Url           string            `gorm:"size:155;not null;"`
  UrlSha1       string            `gorm:"size:40;not null;index"`
  ExtractResult datatypes.JSONMap `gorm:"not null"`
  Status        int               `gorm:"not null;index"`
  CreatedAt     time.Time         `gorm:"not null"`
  UpdatedAt     time.Time         `gorm:"not null"`
}

func (m *Task) TableName() string {
  return "spiders_tasks"
}
