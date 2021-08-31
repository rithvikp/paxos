package paxos

import "time"

const (
	// The amount of time to wait between loop iterations
	loopWaitTime time.Duration = 5 * time.Millisecond
)

type System struct {
	Proposers []*Proposer
	Acceptors []*Acceptor
	Channels  []*Channel
}

func Configure(proposers, acceptors, learners int) *System {
	s := System{}

	proposerInputsFromAcceptors := map[int]*Channel{}
	acceptorInputsFromProposers := []*Channel{}
	// First create relevant channels
	for i := 0; i < proposers; i++ {
		ch := NewChannel()
		proposerInputsFromAcceptors[i] = ch
		s.Channels = append(s.Channels, ch)
	}
	for i := 0; i < acceptors; i++ {
		ch := NewChannel()
		acceptorInputsFromProposers = append(acceptorInputsFromProposers, ch)
		s.Channels = append(s.Channels, ch)
	}

	for i := 0; i < proposers; i++ {
		s.Proposers = append(s.Proposers, &Proposer{
			ID:            i,
			ClientInput:   make(chan int, 10),
			AcceptorInput: proposerInputsFromAcceptors[i],
			Acceptors:     acceptorInputsFromProposers,
		})
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
