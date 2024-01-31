package main

import (
	"github.com/peter-mount/go-kernel/v2"
	"github.com/peter-mount/objectstore"
	"log"
)

func main() {
	err := kernel.Launch(&kernel.MemUsage{}, &objectstore.ObjectStore{})
	if err != nil {
		log.Fatal(err)
	}
}
