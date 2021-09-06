package paxos

import (
	"fmt"
	"math/rand"
	"time"
)

// The supervisor is not currently being used.
type ChannelSupervisor struct{}

func NewChannelSupervisor() *ChannelSupervisor {
	return &ChannelSupervisor{}
}

// Channel defines an non-FIFO, lossy channel. Specifically, messages sent through a Channel could
// be delayed, re-ordered, or lost. The "lossy-ness" of the channel can be configured.
type Channel struct {
	s        *ChannelSupervisor
	rand     *rand.Rand
	read     chan Msg
	write    chan Msg
	dropRate float64
}

// NewChannel creates a new channel. DropRate should be a number between 0 and 1 (inclusive): it
// determines what percentage of messages on the channel will be dropped.
func (s *ChannelSupervisor) NewChannel(dropRate float64) *Channel {
	return &Channel{
		s:        s,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		read:     make(chan Msg, 10),
		write:    make(chan Msg, 10),
		dropRate: dropRate,
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
			time.Sleep(time.Duration(c.rand.Intn(20)) * time.Millisecond)
			if c.rand.Float64() > c.dropRate {
				fmt.Printf("SENT: %s\n", MsgToString(in))
				c.read <- in
			} else {
				fmt.Printf("\tNOT SENT: %s\n", MsgToString(in))
			}
		}
	}
}
