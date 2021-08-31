package paxos

import (
	"fmt"
	"time"
)

type Proposer struct {
	ID            int
	ClientInput   chan int
	AcceptorInput <-chan Msg
	Acceptors     []chan<- Msg

	slot                          int
	smallestAvailableSlot         int
	proposalNumber                int
	highestAcceptedProposalNumber int
	value                         int
	// The number of acceptors which have promised to honor this proposer's proposal.
	promiseCount int
	// Whether accept messages have been sent for the current proposal.
	sentAcceptMsg bool
}

func (p *Proposer) Run() {
	for {
		select {
		case v := <-p.ClientInput:
			p.handleClientInput(v)

		case msg := <-p.AcceptorInput:
			if msg.Promise != nil {
				p.handlePromise(*msg.Promise)
				if p.promiseCount >= len(p.Acceptors)/2+1 && !p.sentAcceptMsg {
					p.sentAcceptMsg = true
					out := Msg{
						Accept: &AcceptMsg{
							Slot:           p.slot,
							ProposerID:     p.ID,
							ProposalNumber: p.proposalNumber,
							Value:          p.value,
						},
					}
					for _, a := range p.Acceptors {
						a <- out
					}
				}
			} else if msg.Accepted != nil {
				// TODO(rithvikp): Potentially check for a majority of acceptances before giving up
				// on using the slot.
				if msg.Accepted.Slot+1 > p.smallestAvailableSlot {
					p.smallestAvailableSlot = msg.Accepted.Slot + 1
				}
			} else {
				fmt.Printf("Proposer %v received a message without any content from acceptor\n", p.ID)
			}
		}
		time.Sleep(loopWaitTime)
	}
}

func (p *Proposer) handleClientInput(val int) {
	p.value = val
	p.proposalNumber++
	p.promiseCount = 0
	p.sentAcceptMsg = false
	p.slot = p.smallestAvailableSlot

	out := Msg{
		Prepare: &PrepareMsg{
			Slot:           p.slot,
			ProposerID:     p.ID,
			ProposalNumber: p.proposalNumber,
		},
	}

	for _, a := range p.Acceptors {
		a <- out
	}
}

func (p *Proposer) handlePromise(msg PromiseMsg) {
	p.promiseCount++

	if msg.AcceptedProposalNumber == nil {
		return
	}

	if *msg.AcceptedProposalNumber > p.highestAcceptedProposalNumber {
		p.value = *msg.AcceptedValue
		p.highestAcceptedProposalNumber = *msg.AcceptedProposalNumber
	}
}
