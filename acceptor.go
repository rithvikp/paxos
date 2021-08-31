package paxos

import (
	"fmt"
	"time"
)

type Acceptor struct {
	ID             int
	ProposerInput  <-chan AcceptorMsg
	ProposerOutput map[int]chan<- ProposerMsg

	// True iff the acceptor should simulate a complete failure.
	CompleteFailure bool

	highestProposalNumber  int
	acceptedProposalNumber *int
	acceptedValue          *int
}

func (a *Acceptor) Run() {
	for {
		if a.CompleteFailure {
			time.Sleep(30 * time.Millisecond)
			continue
		}
		msg := <-a.ProposerInput

		if msg.Prepare != nil {
			a.handlePrepare(*msg.Prepare)
		} else if msg.Accept != nil {
			a.handleAccept(*msg.Accept)
		} else {
			fmt.Printf("Acceptor %v received a message without any content from proposer", a.ID)
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (a *Acceptor) handlePrepare(msg PrepareMsg) {
	if a.highestProposalNumber >= msg.ProposalNumber {
		return
	}
	a.highestProposalNumber = msg.ProposalNumber

	out := PromiseMsg{
		AcceptedProposalNumber: a.acceptedProposalNumber,
		AcceptedValue:          a.acceptedValue,
	}
	a.ProposerOutput[msg.ProposerID] <- ProposerMsg{Promise: &out}
}

func (a *Acceptor) handleAccept(msg AcceptMsg) {
	if a.highestProposalNumber > msg.ProposalNumber {
		return
	}

	a.acceptedProposalNumber = &msg.ProposalNumber
	a.acceptedValue = &msg.Value

	out := AcceptedMsg{AcceptorID: a.ID}
	a.ProposerOutput[msg.ProposerID] <- ProposerMsg{Accepted: &out}
}
