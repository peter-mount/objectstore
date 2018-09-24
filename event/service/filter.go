package service

import (
  "github.com/peter-mount/objectstore/event"
  "strings"
)

// Filter config
type Filter struct {
  S3Key         S3Key       `json:"S3Key" xml:"S3Key" yaml:"S3Key"`
}

type S3Key struct {
  FilterRule  []FilterRule  `json:"FilterRule" xml:"FilterRule" yaml:"FilterRule"`
}

type FilterRule struct {
  Name          string      `json:"Name" xml:"Name" yaml:"Name"`
  Value         string      `json:"Value" xml:"Value" yaml:"Value"`
}

func (f *Filter) Test( evt *event.Event ) bool {
  if len( f.S3Key.FilterRule ) > 0 {
    if evt.S3 == nil || evt.S3.Object == nil {
      return false
    }

    for _, fr := range f.S3Key.FilterRule {
      if !fr.Test( evt ) {
        return false
      }
    }
  }

  return true
}

func (f *FilterRule) Test( evt *event.Event ) bool {
  if f.Name == "prefix" {
    return strings.HasPrefix( evt.S3.Object.Key, f.Value )
  }
  if f.Name == "suffix" {
    return strings.HasSuffix( evt.S3.Object.Key, f.Value )
  }
  return false
}
