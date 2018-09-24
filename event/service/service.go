package service

import (
  "encoding/json"
  "flag"
  "fmt"
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/golib/rabbitmq"
//  "github.com/peter-mount/golib/rest"
  "github.com/peter-mount/objectstore/event"
//  "github.com/peter-mount/objectstore/utils"
  "github.com/streadway/amqp"
  "log"
  "os"
)

// Event service for publishers
type EventService struct {
  events    chan *event.Event
  mq        rabbitmq.RabbitMQ
  pubchan  *amqp.Channel
  url      *string
  exchange *string
  conName  *string
}

func (a *EventService) Name() string {
  return "aws:EventService"
}

func (a *EventService) Init( k *kernel.Kernel ) error {
  a.url = flag.String( "event-url", "", "RabbitMQ connection URL")
  a.exchange = flag.String( "event-exchange", "", "RabbitMQ topic")
  a.conName = flag.String( "event-conname", "", "RabbitMQ connection name")
  return nil
}

func (a *EventService) PostInit() error {
  if *a.url == "" {
    *a.url = os.Getenv( "RABBITMQ_URL" )
  }
  if *a.url != "" {
    a.mq.Url = *a.url

    if *a.exchange == "" {
      *a.exchange = os.Getenv( "RABBITMQ_EXCHANGE" )
    }
    if *a.exchange != "" {
      a.mq.Exchange = *a.exchange
    }

    if *a.conName == "" {
      *a.conName = os.Getenv( "RABBITMQ_CONNAME" )
    }
    if *a.conName != "" {
      a.mq.ConnectionName = *a.conName
    }
  }

  return nil
}

func (a *EventService) Start() error {
  if a.mq.Url != "" {
    err := a.mq.Connect()
    if err != nil {
      return err
    }

    //a.pubchan, err = a.mq.NewChannel()
    //if err != nil {
    //  return err
    //}
  }

  a.events = make( chan *event.Event, 20 )
  go a.publisher()

  return nil
}

// Notify accepts Events for submission
func (a *EventService) Notify( evt *event.Event ) {
  a.events <- evt
}

func (a *EventService) publisher() {
  for {
    evt := <-a.events

    // Ensure event is complete
    if evt.Version == "" {
      evt.Version = "2.0"
    }
    if evt.S3 != nil && evt.S3.Version == "" {
      evt.S3.Version = "1.0"
    }

    b, err := json.MarshalIndent( evt, "", "  " )
    if err != nil {
      log.Println( err )
    }
    log.Printf( "Event:\n %s", string(b[:]) )

    // TODO filter events here
    err = a.publishEvent( evt )
    if err != nil {
      log.Println( "Failed to publish event", err )
    }
  }
}

func (a *EventService) publishEvent( evt *event.Event ) error {
  b, err := json.Marshal( &event.Records{ []*event.Event{ evt } } )
  if err != nil {
    return err
  }

  // The routing key, this is an ARN
  var routingKey string

  if evt.S3 != nil {
    // S3 routing key
    // source should be "aws:s3" and region the bucket region. No user id then
    // the bucket/key as the resource
    routingKey = fmt.Sprintf(
      "arn:%s:%s::%s/%s",
      evt.Source,
      evt.Region,
      evt.S3.Bucket.Name,
      evt.S3.Object.Key,
    )
  } else {
    // Basic routing key of source service & region -fail safe option
    routingKey = fmt.Sprintf(
      "arn:%s:%s:::",
      evt.Source,
      evt.Region,
    )
  }
  return a.mq.Publish( routingKey, b )
}
