package service

import (
  "encoding/json"
  "errors"
  "github.com/peter-mount/golib/rabbitmq"
  "github.com/peter-mount/objectstore/event"
)

// Our extension, a RabbitMQ connection
type RabbitConfiguration struct {
  Id        string      `json:"Id" xml:"Id" yaml:"Id"`
  // The AmqpUrl to use
  AmqpUrl   string      `json:"Amqp" xml:"Amqp" yaml:"Amqp"`
  // The exchange to use
  Exchange  string      `json:"Exchange" xml:"Exchange" yaml:"Exchange"`

  Event     string      `json:"Event" xml:"Event" yaml:"Event"`
  Filter    Filter      `json:"Filter" xml:"Filter" yaml:"Filter"`
}

type RabbitMQ struct {
  mq            rabbitmq.RabbitMQ
  // Config for this instance
  config     []*RabbitConfiguration
}

func (a *EventService) loadRabbitConfig( conf []*RabbitConfiguration ) error {
  for _, cfg := range conf {
    if cfg.AmqpUrl == "" {
      return errors.New( "Amqp url is mandatory" )
    }
    if cfg.Exchange == "" {
      cfg.Exchange = "amq.topic"
    }

    key := cfg.Exchange + ":" + cfg.AmqpUrl
    mq, exists := a.mqInstances[ key ]
    if !exists {
      mq = &RabbitMQ{}

      mq.mq.Url = cfg.AmqpUrl
      mq.mq.Exchange = cfg.Exchange
      mq.mq.ConnectionName = "ObjectStore event publisher " + cfg.Exchange

      a.mqInstances[ key ] = mq
    }

    mq.config = append( mq.config, cfg )
  }

  return nil
}

func (a *EventService) startRabbitMQ() error {
  for _, mq := range a.mqInstances {
    err := mq.mq.Connect()
    if err != nil {
      return err
    }
  }

  return nil
}

func (a *EventService) publishRabbit( evt *event.Event ) error {
  var b []byte
  var err error

  found := false
  for _, c := range a.mqInstances {

    if !found {
      found = true
      b, err = json.Marshal( &event.Records{ []*event.Event{ evt } } )
      if err != nil {
        return err
      }
    }

    c.mq.Publish( evt.RoutingKey(), b )
  }

  return nil
}
