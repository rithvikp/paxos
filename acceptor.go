package paxos

import (
	"fmt"
	"time"
)

type slotState struct {
	slot                   int
	highestProposalNumber  int
	acceptedProposalNumber *int
	acceptedValue          *int
}

type Acceptor struct {
	ID             int
	ProposerInput  <-chan Msg
	ProposerOutput map[int]chan<- Msg

	// True iff the acceptor should simulate a complete failure.
	CompleteFailure bool

	state map[int]*slotState
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

		time.Sleep(loopWaitTime)
	}
}

func (a *Acceptor) handlePrepare(msg PrepareMsg) {
	if _, ok := a.state[msg.Slot]; !ok {
		a.state[msg.Slot] = &slotState{
			slot:                  msg.Slot,
			highestProposalNumber: -1,
		}
	}
	state := a.state[msg.Slot]

	if state.highestProposalNumber >= msg.ProposalNumber {
		return
	}
	state.highestProposalNumber = msg.ProposalNumber

	out := PromiseMsg{
		Slot:                   state.slot,
		AcceptedProposalNumber: state.acceptedProposalNumber,
		AcceptedValue:          state.acceptedValue,
	}
	a.ProposerOutput[msg.ProposerID] <- Msg{Promise: &out}
}

func (a *Acceptor) handleAccept(msg AcceptMsg) {
	// TODO(rithvikp): Should this situation be an error instead?
	if _, ok := a.state[msg.Slot]; !ok {
		a.state[msg.Slot] = &slotState{
			slot:                  msg.Slot,
			highestProposalNumber: -1,
		}
	}
	state := a.state[msg.Slot]

	if state.highestProposalNumber > msg.ProposalNumber {
		return
	}

	// TODO(rithvikp): Should there be an error if this value is different from a previously accepted value?

	state.acceptedProposalNumber = &msg.ProposalNumber
	state.acceptedValue = &msg.Value

	out := AcceptedMsg{
		Slot:           state.slot,
		AcceptorID:     a.ID,
		ProposalNumber: msg.ProposalNumber,
		Value:          msg.Value,
	}

	for _, ch := range a.ProposerOutput {
		ch <- Msg{Accepted: &out}
	}

	fmt.Printf("Accepted a value for slot %d --> Proposer: %v, Acceptor: %v, Proposal #: %d, Value: %v\n",
		state.slot, msg.ProposerID, a.ID, msg.ProposalNumber, msg.Value)
}
