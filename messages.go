package paxos

import (
	"encoding/json"
	"fmt"
)

// Client request
type ClientMsg struct {
	Value int
}

// Leader updates
type LeaderMsg struct {
	ProposerID     int
	DestProposerID int
}

// Phase 1A
type PrepareMsg struct {
	Slot           int
	ProposerID     int
	AcceptorID     int
	ProposalNumber int
}

// Phase 1B
type PromiseMsg struct {
	Slot                   int
	ProposerID             int
	AcceptorID             int
	AcceptedProposalNumber *int
	AcceptedValue          *int
}

// Phase 2A
type AcceptMsg struct {
	Slot           int
	ProposerID     int
	AcceptorID     int
	ProposalNumber int
	Value          int
}

// Phase 2B
type AcceptedMsg struct {
	Slot           int
	ProposerID     int
	AcceptorID     int
	ProposalNumber int
	Value          int
}

type LearnMsg struct {
	Slot       int
	AcceptorID int
	LearnerID  int
	Value      int
}

// A wrapper that is uses when messages are passed around.
type Msg struct {
	Client   *ClientMsg
	Leader   *LeaderMsg
	Prepare  *PrepareMsg
	Accept   *AcceptMsg
	Promise  *PromiseMsg
	Accepted *AcceptedMsg
	Learn    *LearnMsg
}

func MsgToString(msg Msg) string {
	var data interface{}
	var prefix string
	if msg.Client != nil {
		data = msg.Client
		prefix = "Client"
	} else if msg.Leader != nil {
		data = msg.Leader
		prefix = "Leader"
	} else if msg.Prepare != nil {
		data = msg.Prepare
		prefix = "Prepare"
	} else if msg.Accept != nil {
		data = msg.Accept
		prefix = "Accept"
	} else if msg.Promise != nil {
		data = msg.Promise
		prefix = "Promise"
	} else if msg.Accepted != nil {
		data = msg.Accepted
		prefix = "Accepted"
	} else if msg.Learn != nil {
		data = msg.Learn
		prefix = "Learn"
	}

	b, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("Unable to convert the given message to a string: %v", err))
	}

	return fmt.Sprintf("%s: %s", prefix, b)
}
