package service

import (
  "flag"
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/objectstore/event"
)

// Event service for publishers
type EventService struct {
  events        chan *event.Event

  mqInstances   map[string]*RabbitMQ
  configFile   *string
}

func (a *EventService) Name() string {
  return "aws:EventService"
}

func (a *EventService) Init( k *kernel.Kernel ) error {
  a.configFile = flag.String( "event-config", "", "Event configuration")
  return nil
}

func (a *EventService) PostInit() error {
  if err := a.loadConfig(); err != nil {
    return err
  }

  return nil
}

func (a *EventService) Start() error {
  // Start each broker
  err := a.startRabbitMQ()
  if err != nil {
    return err
  }

  // Create the channel & the publisher thread
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

    // Pass to each possible broker here
    a.publishRabbit( evt )
  }
}
