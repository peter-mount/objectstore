package service

import (
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "log"
  "path/filepath"
)

// Notification configuration, loosely based on aws
// https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-filtering
// with additions to support
type NotificationConfiguration struct {
  RabbitConfig []*RabbitConfiguration      `json:"RabbitConfig" xml:"RabbitConfig" yaml:"RabbitConfig"`
}

// Filter config
type Filter struct {
  S3Key         S3Key       `json:"S3Key" xml:"S3Key" yaml:"S3Key"`
  FilterRule  []FilterRule  `json:"FilterRule" xml:"FilterRule" yaml:"FilterRule"`
}
type S3Key struct {
  FilterRule  []FilterRule  `json:"FilterRule" xml:"FilterRule" yaml:"FilterRule"`
}
type FilterRule struct {
  Name          string      `json:"Name" xml:"Name" yaml:"Name"`
  Value         string      `json:"Value" xml:"Value" yaml:"Value"`
}

func (a *EventService) loadConfig() error {
  a.mqInstances = make( map[string]*RabbitMQ )

  if *a.configFile == "" {
    return nil
  }

  filename, _ := filepath.Abs( *a.configFile )
  log.Println( "Loading event config:", filename )

  yml, err := ioutil.ReadFile( filename )
  if err != nil {
    return err
  }

  config := &NotificationConfiguration{}
  err = yaml.Unmarshal( yml, config )
  if err != nil {
    return err
  }

  if config.RabbitConfig != nil {
    err = a.loadRabbitConfig( config.RabbitConfig )
    if err != nil {
      return err
    }
  }

  return nil
}
