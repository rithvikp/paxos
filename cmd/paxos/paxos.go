package main

import (
	"time"

	"github.com/rithvikp/paxos"
)

const (
	f int = 3
)

func main() {
	s := paxos.Configure(2, 5, 0)

	s.Run()

	s.Acceptors[0].CompleteFailure = true

	s.Proposers[0].ClientInput <- 2
	//time.Sleep(1 * time.Second)
	s.Proposers[1].ClientInput <- 3

	for {
		time.Sleep(200 * time.Second)
	}
}
