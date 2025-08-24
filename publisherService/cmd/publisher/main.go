package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func main(){

  fmt.Println("publisher service starting")

  forever := make(chan bool)
  go func() {

    for i := 0; true; i++ {
      id := uuid.New()
      fmt.Printf("publisher iteration %d. GUID = %s\n", i, &id)
      time.Sleep(2 * time.Second)
    }

  }()
  <-forever
}
