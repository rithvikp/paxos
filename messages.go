package paxos

// Phase 1A
type PrepareMsg struct {
	Slot           int
	ProposerID     int
	ProposalNumber int
}

// Phase 1B
type PromiseMsg struct {
	Slot                   int
	AcceptedProposalNumber *int
	AcceptedValue          *int
}

// Phase 2A
type AcceptMsg struct {
	Slot           int
	ProposerID     int
	ProposalNumber int
	Value          int
}

// Phase 2B
type AcceptedMsg struct {
	Slot           int
	AcceptorID     int
	ProposalNumber int
	Value          int
}

// A wrapper that is uses when messages are passed around.
type Msg struct {
	Prepare  *PrepareMsg
	Accept   *AcceptMsg
	Promise  *PromiseMsg
	Accepted *AcceptedMsg
}
