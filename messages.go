package paxos

// Phase 1A
type PrepareMsg struct {
	ProposerID     int
	ProposalNumber int
}

// Phase 1B
type PromiseMsg struct {
	AcceptedProposalNumber *int
	AcceptedValue          *int
}

// Phase 2A
type AcceptMsg struct {
	ProposerID     int
	ProposalNumber int
	Value          int
}

// Phase 2B
type AcceptedMsg struct {
	AcceptorID int
}

// Phase 2B
type LearnerAcceptedMsg struct {
	ProposalNumber int
	Value          int
}

// Messages to send to acceptors.
type AcceptorMsg struct {
	Prepare *PrepareMsg
	Accept  *AcceptMsg
}

// Messages to send to proposers.
type ProposerMsg struct {
	Promise  *PromiseMsg
	Accepted *AcceptedMsg
}
