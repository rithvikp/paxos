package paxos

import "time"

const (
	// The amount of time to wait between loop iterations
	loopWaitTime time.Duration = 1 * time.Millisecond
)

type System struct {
	Proposers []*Proposer
	Acceptors []*Acceptor
	Channels  []*Channel
}

func Configure(proposers, acceptors, learners int) *System {
	s := System{}

	proposerInputsFromAcceptors := map[int]*Channel{}
	acceptorInputsFromProposers := map[int]*Channel{}

	supervisor := NewChannelSupervisor()

	// First create relevant channels
	for i := 0; i < proposers; i++ {
		ch := supervisor.NewChannel()
		proposerInputsFromAcceptors[i] = ch
		s.Channels = append(s.Channels, ch)
	}
	for i := 0; i < acceptors; i++ {
		ch := supervisor.NewChannel()
		acceptorInputsFromProposers[i] = ch
		s.Channels = append(s.Channels, ch)
	}

	for i := 0; i < proposers; i++ {
		p := &Proposer{
			ID:            i,
			ClientInput:   make(chan int, 10),
			AcceptorInput: proposerInputsFromAcceptors[i],
			Acceptors:     acceptorInputsFromProposers,
		}
		p.Init()

		s.Proposers = append(s.Proposers, p)
	}

	for i := 0; i < acceptors; i++ {
		s.Acceptors = append(s.Acceptors, &Acceptor{
			ID:             i,
			ProposerInput:  acceptorInputsFromProposers[i],
			ProposerOutput: proposerInputsFromAcceptors,
			state:          map[int]*slotState{},
		})
	}

	return &s
}

func (s *System) Run() {
	for _, p := range s.Proposers {
		go p.Run()
	}

	for _, a := range s.Acceptors {
		go a.Run()
	}

	for _, ch := range s.Channels {
		go ch.Run()
	}
}
