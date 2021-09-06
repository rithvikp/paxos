package paxos

import "time"

const (
	// The amount of time to wait between loop iterations
	loopWaitTime time.Duration = 1 * time.Millisecond
)

type System struct {
	Proposers []*Proposer
	Acceptors []*Acceptor
	Learners  []*Learner
	Channels  []*Channel
}

func Configure(proposers, acceptors, learners int) *System {
	s := System{}

	proposerInputsFromClients := map[int]chan int{}
	heartbeatInputFromProposers := map[int]*Channel{}
	proposerInputsFromAcceptors := map[int]*Channel{}
	acceptorInputsFromProposers := map[int]*Channel{}
	learnerInputFromAcceptors := map[int]*Channel{}

	supervisor := NewChannelSupervisor()

	// First create relevant channels
	for i := 0; i < proposers; i++ {
		ch := supervisor.NewChannel(0.2, true)
		proposerInputsFromAcceptors[i] = ch
		s.Channels = append(s.Channels, ch)

		ch = supervisor.NewChannel(0.2, false)
		heartbeatInputFromProposers[i] = ch
		s.Channels = append(s.Channels, ch)

		clientCh := make(chan int, 10)
		proposerInputsFromClients[i] = clientCh
	}
	for i := 0; i < acceptors; i++ {
		ch := supervisor.NewChannel(0.2, true)
		acceptorInputsFromProposers[i] = ch
		s.Channels = append(s.Channels, ch)
	}
	for i := 0; i < learners; i++ {
		ch := supervisor.NewChannel(0, true)
		learnerInputFromAcceptors[i] = ch
		s.Channels = append(s.Channels, ch)
	}

	for i := 0; i < proposers; i++ {
		p := &Proposer{
			ID:                 i,
			ClientInput:        proposerInputsFromClients[i],
			Proposers:          proposerInputsFromClients,
			HeartbeatInput:     heartbeatInputFromProposers[i],
			ProposerHeartbeats: heartbeatInputFromProposers,
			AcceptorInput:      proposerInputsFromAcceptors[i],
			Acceptors:          acceptorInputsFromProposers,
		}
		p.Init()

		s.Proposers = append(s.Proposers, p)
	}

	for i := 0; i < acceptors; i++ {
		s.Acceptors = append(s.Acceptors, &Acceptor{
			ID:            i,
			ProposerInput: acceptorInputsFromProposers[i],
			Proposers:     proposerInputsFromAcceptors,
			Learners:      learnerInputFromAcceptors,
			state:         map[int]*slotState{},
		})
	}

	for i := 0; i < learners; i++ {
		s.Learners = append(s.Learners, &Learner{
			ID:            i,
			AcceptorInput: learnerInputFromAcceptors[i],
			AcceptorCount: acceptors,
			Log:           map[int]int{},
			PendingLogs:   map[int]*pendingLogEntry{},
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

	for _, l := range s.Learners {
		go l.Run()
	}

	for _, ch := range s.Channels {
		go ch.Run()
	}
}
