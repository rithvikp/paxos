package paxos

import (
	"fmt"
	"sync"
	"time"
)

const (
	leaderUpdateInterval time.Duration = 100 * time.Millisecond
)

type leaderInfo struct {
	leaderID           int
	updateLeaderToSelf bool
}

type acceptorState struct {
	Channel             *Channel
	HighestAcceptedSlot int
}

type Proposer struct {
	ID                 int
	ClientInput        chan int
	Proposers          map[int]chan int
	HeartbeatInput     *Channel         // For leader selection.
	ProposerHeartbeats map[int]*Channel // For leader selection.
	AcceptorInput      *Channel
	Acceptors          map[int]*acceptorState

	leaderInfoMu sync.RWMutex
	leaderInfo   leaderInfo

	slot                          int
	value                         int
	proposalNumber                int
	highestAcceptedProposalNumber int
	// The number of acceptors which have promised to honor this proposer's proposal.
	promiseCount int
	// Whether accept messages have been sent for the current proposal.
	sentAcceptMsg bool
	skipPrepare   bool

	resendTimer *time.Timer

	smallestAvailableSlot int
}

const (
	proposalRepeatInterval time.Duration = 1 * time.Second
)

func (p *Proposer) Init() {
	p.highestAcceptedProposalNumber = -1
	p.leaderInfo.leaderID = p.ID

	timer := time.NewTimer(proposalRepeatInterval)
	timer.Stop()
	p.resendTimer = timer
}

func (p *Proposer) Run() {
	go p.trackLeader()
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
				if msg.Accepted.ProposalNumber == p.proposalNumber {
					p.resendTimer.Stop()
				}
			} else {
				fmt.Printf("Proposer %v received a message without any content from acceptor\n", p.ID)
			}

		case <-p.resendTimer.C:
			//fmt.Printf("Resend, proposer %v, proposal # %d\n", p.ID, p.proposalNumber)
			p.handleClientInput(p.value)

		case v := <-p.ClientInput:
			if p.isLeader() {
				if p.skipPrepare {
				} else {
					p.handleClientInput(v)
				}
			} else {
				leader := p.leader()
				fmt.Printf("[PROPOSER] Forwarded a message from proposer %v to proposer %v\n", p.ID, leader)
				p.Proposers[leader] <- v
			}
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

	for aID, a := range p.Acceptors {
		out := Msg{
			Prepare: &PrepareMsg{
				Slot:           p.slot,
				ProposerID:     p.ID,
				AcceptorID:     aID,
				ProposalNumber: p.proposalNumber,
			},
		}
		a.Channel.Write() <- out
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
		p.resendTimer = time.NewTimer(proposalRepeatInterval)
		p.sentAcceptMsg = true

		for aID, a := range p.Acceptors {
			out := Msg{
				Accept: &AcceptMsg{
					Slot:           p.slot,
					ProposerID:     p.ID,
					AcceptorID:     aID,
					ProposalNumber: p.proposalNumber,
					Value:          p.value,
				},
			}
			a.Channel.Write() <- out
		}
	}
}

func (p *Proposer) trackLeader() {
	for {
		// Send heartbeat messages
		for destProposerID, ch := range p.ProposerHeartbeats {
			out := LeaderMsg{
				ProposerID:     p.ID,
				DestProposerID: destProposerID,
			}
			ch.Write() <- Msg{Leader: &out}
		}

		// Update the leader based on the received heartbeats
		highest := p.ID
		run := true
		for run {
			select {
			case msg := <-p.HeartbeatInput.Read():
				if msg.Leader != nil {
					if msg.Leader.ProposerID > highest {
						highest = msg.Leader.ProposerID
					}
				} else {
					fmt.Printf("Proposer %v received a message without any relevant content from another proposer\n", p.ID)
				}
			default:
				run = false
			}
		}

		p.leaderInfoMu.Lock()
		if highest > p.ID {
			p.leaderInfo.leaderID = highest
			p.leaderInfo.updateLeaderToSelf = false
		} else if p.leaderInfo.updateLeaderToSelf {
			p.leaderInfo.leaderID = p.ID
			p.leaderInfo.updateLeaderToSelf = false
		} else {
			p.leaderInfo.updateLeaderToSelf = true
		}
		p.leaderInfoMu.Unlock()

		time.Sleep(leaderUpdateInterval)
	}
}

func (p *Proposer) isLeader() bool {
	p.leaderInfoMu.RLock()
	defer p.leaderInfoMu.RUnlock()

	return p.leaderInfo.leaderID == p.ID
}

func (p *Proposer) leader() int {
	p.leaderInfoMu.RLock()
	defer p.leaderInfoMu.RUnlock()

	return p.leaderInfo.leaderID
}
