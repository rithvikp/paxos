package paxos

import (
	"fmt"
	"time"
)

type pendingLogEntry struct {
	// acceptedValuesToCount maps values that were accepted to the number of acceptors that accepted
	// them.
	AcceptedValuesToCount map[int]int
}

type Learner struct {
	ID            int
	AcceptorInput *Channel
	AcceptorCount int
	// The slot number to value.
	Log map[int]int
	// Map slot number to a pending log entry.
	PendingLogs map[int]*pendingLogEntry
}

func (l *Learner) Run() {
	for {
		msg := <-l.AcceptorInput.Read()

		if msg.Learn != nil {
			l.handleLearn(*msg.Learn)
		} else {
			fmt.Printf("Learner %v received a message without any relevant content from acceptor", l.ID)
		}

		time.Sleep(loopWaitTime)
	}
}

func (l *Learner) handleLearn(msg LearnMsg) {
	if _, ok := l.Log[msg.Slot]; ok {
		// TODO(rithvikp): Handle updates
		return
	}

	pl, ok := l.PendingLogs[msg.Slot]
	if !ok {
		pl = &pendingLogEntry{
			AcceptedValuesToCount: map[int]int{},
		}
		l.PendingLogs[msg.Slot] = pl
	}
	pl.AcceptedValuesToCount[msg.Value] += 1

	if pl.AcceptedValuesToCount[msg.Value] >= l.AcceptorCount/2+1 {
		l.Log[msg.Slot] = msg.Value

		fmt.Printf("[LEARNER] Accepted a value for slot %d --> Acceptor: %v, Learner: %v, Value: %v\n", msg.Slot, msg.AcceptorID, l.ID, msg.Value)
	}
}
