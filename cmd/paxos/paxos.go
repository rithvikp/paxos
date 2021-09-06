package main

import (
	"fmt"
	"time"

	"github.com/rithvikp/paxos"
)

const (
	f int = 3
)

func main() {
	fmt.Println("Starting Paxos")
	s := paxos.Configure(2, 5, 1)

	s.Run()

	s.Proposers[0].ClientInput <- 2
	time.Sleep(1000 * time.Millisecond)
	s.Proposers[1].ClientInput <- 3

	for {
		time.Sleep(200 * time.Second)
	}
}
