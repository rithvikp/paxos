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
	ID            int
	ProposerInput *Channel
	Proposers     map[int]*Channel
	Learners      map[int]*Channel

	state map[int]*slotState
}

func (a *Acceptor) Run() {
	for {
		msg := <-a.ProposerInput.Read()

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

	maxAcceptedSlot := -1
	for slot, st := range a.state {
		if st.acceptedValue == nil {
			continue
		}
		if slot > maxAcceptedSlot {
			maxAcceptedSlot = slot
		}
	}

	out := PromiseMsg{
		Slot:                   state.slot,
		ProposerID:             msg.ProposerID,
		AcceptorID:             a.ID,
		AcceptedProposalNumber: state.acceptedProposalNumber,
		AcceptedValue:          state.acceptedValue,
		HighestAcceptedSlot:    maxAcceptedSlot,
	}
	a.Proposers[msg.ProposerID].Write() <- Msg{Promise: &out}
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

	for pID, ch := range a.Proposers {
		out := AcceptedMsg{
			Slot:           state.slot,
			ProposerID:     pID,
			AcceptorID:     a.ID,
			ProposalNumber: msg.ProposalNumber,
			Value:          msg.Value,
		}
		ch.Write() <- Msg{Accepted: &out}
	}

	for lID, ch := range a.Learners {
		out := LearnMsg{
			Slot:       state.slot,
			AcceptorID: a.ID,
			LearnerID:  lID,
			Value:      msg.Value,
		}
		ch.Write() <- Msg{Learn: &out}
	}

	fmt.Printf("[ACCEPTOR] Accepted a value for slot %d --> Proposer: %v, Acceptor: %v, Proposal #: %d, Value: %v\n",
		state.slot, msg.ProposerID, a.ID, msg.ProposalNumber, msg.Value)
}
