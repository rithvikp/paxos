package paxos

import (
	"fmt"
	"time"
)

type Proposer struct {
	ID                            int
	ClientInput                   chan int
	AcceptorInput                 <-chan ProposerMsg
	Acceptors                     []chan<- AcceptorMsg
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
			p.handlePropose(v)

		case msg := <-p.AcceptorInput:
			if msg.Promise != nil {
				p.handlePromise(*msg.Promise)
				if p.promiseCount >= len(p.Acceptors)/2+1 && !p.sentAcceptMsg {
					p.sentAcceptMsg = true
					out := AcceptorMsg{
						Accept: &AcceptMsg{
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
				fmt.Printf("Accepted a value! Proposer %v, Acceptor %v, #: %v, Value: %v\n", p.ID, msg.Accepted.AcceptorID, p.proposalNumber, p.value)
			} else {
				fmt.Printf("Proposer %v received a message without any content from acceptor\n", p.ID)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (p *Proposer) handlePropose(val int) {
	p.value = val
	p.proposalNumber++
	p.promiseCount = 0
	p.sentAcceptMsg = false

	out := AcceptorMsg{
		Prepare: &PrepareMsg{
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
