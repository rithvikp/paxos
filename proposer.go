package paxos

import (
	"fmt"
	"time"
)

type Proposer struct {
	ID            int
	ClientInput   chan int
	AcceptorInput *Channel
	Acceptors     map[int]*Channel

	slot                          int
	value                         int
	proposalNumber                int
	highestAcceptedProposalNumber int
	// The number of acceptors which have promised to honor this proposer's proposal.
	promiseCount int
	// Whether accept messages have been sent for the current proposal.
	sentAcceptMsg bool

	resendTimer *time.Timer

	smallestAvailableSlot int
}

const (
	proposalRepeatInterval time.Duration = 1 * time.Second
)

func (p *Proposer) Init() {
	p.highestAcceptedProposalNumber = -1

	timer := time.NewTimer(1 * time.Second)
	timer.Stop()
	p.resendTimer = timer
}

func (p *Proposer) Run() {
	for {
		select {
		case msg := <-p.AcceptorInput.Read():
			//fmt.Printf("Proposer %v, received msg: %+v\n", p.ID, msg)
			if msg.Promise != nil {
				p.handlePromise(*msg.Promise)
			} else if msg.Accepted != nil {
				// TODO(rithvikp): Potentially check for a majority of acceptances before giving up
				// on using the slot.
				if msg.Accepted.Slot+1 > p.smallestAvailableSlot {
					p.smallestAvailableSlot = msg.Accepted.Slot + 1
				}
			} else {
				fmt.Printf("Proposer %v received a message without any content from acceptor\n", p.ID)
			}

		case <-p.resendTimer.C:
			//fmt.Printf("Resend, proposer %v, proposal # %d\n", p.ID, p.proposalNumber)
			p.handleClientInput(p.value)

		case v := <-p.ClientInput:
			p.proposalNumber = 0
			p.handleClientInput(v)
		}

		time.Sleep(loopWaitTime)
	}
}

func (p *Proposer) handleClientInput(val int) {
	p.slot = p.smallestAvailableSlot
	p.value = val
	p.proposalNumber++

	p.highestAcceptedProposalNumber = -1
	p.promiseCount = 0
	p.sentAcceptMsg = false
	p.resendTimer = time.NewTimer(proposalRepeatInterval)

	for id, a := range p.Acceptors {
		out := Msg{
			Prepare: &PrepareMsg{
				Slot:           p.slot,
				ProposerID:     p.ID,
				AcceptorID:     id,
				ProposalNumber: p.proposalNumber,
			},
		}
		a.Write() <- out
	}
}

func (p *Proposer) handlePromise(msg PromiseMsg) {
	p.promiseCount++

	if msg.AcceptedProposalNumber != nil && *msg.AcceptedProposalNumber > p.highestAcceptedProposalNumber {
		p.value = *msg.AcceptedValue
		p.highestAcceptedProposalNumber = *msg.AcceptedProposalNumber
	}

	// Enter phase 2 if applicable.
	if p.promiseCount >= len(p.Acceptors)/2+1 && !p.sentAcceptMsg {
		p.resendTimer.Stop()
		p.sentAcceptMsg = true

		for id, a := range p.Acceptors {
			out := Msg{
				Accept: &AcceptMsg{
					Slot:           p.slot,
					ProposerID:     p.ID,
					AcceptorID:     id,
					ProposalNumber: p.proposalNumber,
					Value:          p.value,
				},
			}
			a.Write() <- out
		}
	}
}
