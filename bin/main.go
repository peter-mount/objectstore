package main

import (
  "github.com/peter-mount/golib/kernel"
  "github.com/peter-mount/objectstore"
  "log"
)

func main() {
  err := kernel.Launch( &kernel.MemUsage{}, &objectstore.ObjectStore{} )
  if err != nil {
    log.Fatal( err )
  }
}
