package raft

import (
	"time"
)

type State string

const (
	Follower           State = "follower"
	Candidate          State = "candidate"
	Leader             State = "leader"
	ElectionTimeoutMax       = 400
	ElectionTimeoutMin       = 200
	HeartBeatTimeOut         = 50
)

func (rf *Raft) becomeLeader() {
	rf.state = Leader
	//rf.votedFor = -1
	rf.getVoteNum = 0
	rf.lastResetElectionTime = time.Now()
}
func (rf *Raft) becomeCandidate() {
	rf.state = Candidate
	rf.votedFor = rf.me // vote for me
	rf.getVoteNum = 1
	rf.currentTerm += 1
	//fmt.Printf("rf.me:%d----rf.currentTerm: %d-----rf..votedFor: %d\n", rf.me, rf.currentTerm, rf.votedFor)
	rf.JoinElection()
	rf.lastResetElectionTime = time.Now()
}
func (rf *Raft) becomeFollower(Term int) {
	rf.state = Follower
	rf.currentTerm = Term
	rf.votedFor = -1
	rf.getVoteNum = 0
}
