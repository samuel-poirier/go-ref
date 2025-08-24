package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func main(){

  fmt.Println("consumer service starting")

  forever := make(chan bool)
  go func() {

    for i := 0; true; i++ {
      id := uuid.New()
      fmt.Printf("consumer iteration %d. GUID = %s\n", i, &id)
      time.Sleep(2 * time.Second)
    }

  }()
  <-forever
}
