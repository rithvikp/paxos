package paxos

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type ChannelSupervisor struct {
	mu     *sync.Mutex
	reader *bufio.Reader
}

func NewChannelSupervisor() *ChannelSupervisor {
	return &ChannelSupervisor{
		mu:     &sync.Mutex{},
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *ChannelSupervisor) AskToSend(data interface{}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.MarshalIndent(data, "", "\t")
	fmt.Printf("%s\nAccept?: ", b)
	text, _ := s.reader.ReadString('\n')
	return text == "y"
}

// Channel defines an non-FIFO, lossy channel. Specifically, messages sent through a Channel could
// be delayed, re-ordered, or lost. The "lossy-ness" of the channel can be configured.
type Channel struct {
	s     *ChannelSupervisor
	read  chan Msg
	write chan Msg
}

func (s *ChannelSupervisor) NewChannel() *Channel {
	return &Channel{
		s:     s,
		read:  make(chan Msg, 10),
		write: make(chan Msg, 10),
	}
}

func (c *Channel) Read() <-chan Msg {
	return c.read
}

func (c *Channel) Write() chan<- Msg {
	return c.write
}

func (c *Channel) Run() {
	for {
		select {
		case in := <-c.write:
			if c.s.AskToSend(in) {
				c.read <- in
			}
		}
	}
}
