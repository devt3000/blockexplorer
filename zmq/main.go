package main

import (
	"github.com/zeromq/goczmq"
	"log"
	"encoding/hex"
	"fmt"
)

func main() {

	subscriber, err := goczmq.NewSub("tcp://127.0.0.1:28336", "hashblock")
	if err != nil {
		log.Fatal(err)
	}

	defer subscriber.Destroy()

	for {
		msg, _, err := subscriber.RecvFrame()

		if err != nil {
			log.Fatalf("panic reeeeee %s", err)
		}
		fmt.Println(hex.Dump(msg))
	}

}