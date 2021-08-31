package paxos

// Channel defines an non-FIFO, lossy channel. Specifically, messages sent through a Channel could
// be delayed, re-ordered, or lost. The "lossy-ness" of the channel can be configured.
type Channel struct {
	read  chan Msg
	write chan Msg
}

func NewChannel() *Channel {
	return &Channel{
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
			c.read <- in
		}
	}
}
