package paxos

import "time"

const (
	// The amount of time to wait between loop iterations
	loopWaitTime time.Duration = 5 * time.Millisecond
)

type System struct {
	Proposers []*Proposer
	Acceptors []*Acceptor
}

func Configure(proposers, acceptors, learners int) *System {
	s := System{}

	proposerInputsFromAcceptors := map[int]chan Msg{}
	proposerInputsFromAcceptorsWO := map[int]chan<- Msg{}
	acceptorInputsFromProposers := []chan Msg{}
	acceptorInputsFromProposersWO := []chan<- Msg{}
	// First create relevant channels
	for i := 0; i < proposers; i++ {
		ch := make(chan Msg, 10)
		proposerInputsFromAcceptors[i] = ch
		proposerInputsFromAcceptorsWO[i] = ch
	}
	for i := 0; i < acceptors; i++ {
		ch := make(chan Msg, 10)
		acceptorInputsFromProposers = append(acceptorInputsFromProposers, ch)
		acceptorInputsFromProposersWO = append(acceptorInputsFromProposersWO, ch)
	}

	for i := 0; i < proposers; i++ {
		s.Proposers = append(s.Proposers, &Proposer{
			ID:            i,
			ClientInput:   make(chan int, 10),
			AcceptorInput: proposerInputsFromAcceptors[i],
			Acceptors:     acceptorInputsFromProposersWO,
		})
	}

	for i := 0; i < acceptors; i++ {
		s.Acceptors = append(s.Acceptors, &Acceptor{
			ID:             i,
			ProposerInput:  acceptorInputsFromProposers[i],
			ProposerOutput: proposerInputsFromAcceptorsWO,
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
}
