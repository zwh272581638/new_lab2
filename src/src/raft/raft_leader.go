package raft

import (
	"fmt"
	"time"
)

func (rf *Raft) JoinElection() {
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}

		go func(server int) {
			rf.mu.Lock()
			args := RequestVoteArgs{
				rf.currentTerm,
				rf.me,
				0,
				0,
			}
			voteReply := RequestVoteReply{}
			rf.mu.Unlock()

			// waiting code should free lock first.
			ok := rf.sendRequestVote(server, &args, &voteReply)
			if !ok {
				return
			}

			rf.mu.Lock()
			defer rf.mu.Unlock()
			if voteReply.Term > args.Term {
				rf.becomeFollower(voteReply.Term)
				return
			}
			if !voteReply.VoteGranted {
				return
			}
			//if rf.state != Candidate || args.Term != rf.currentTerm {
			//	return
			//}

			if rf.state == Candidate && voteReply.VoteGranted == true && rf.currentTerm == args.Term {
				rf.getVoteNum += 1
				if rf.getVoteNum >= len(rf.peers)/2+1 {
					fmt.Printf("%d成为了leader\n", rf.me)
					rf.becomeLeader()
					rf.leaderSendEntries()
					return
				}
			}

		}(peer)
	}
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// All servers 2原则，candidate任期大则投票的raft节点必转为follower，然后Term相同
	if args.Term > rf.currentTerm {
		rf.becomeFollower(args.Term)
	}
	// 此时rf.currentTerm <= args.Term
	// request vote rpc receiver 1
	if args.Term < rf.currentTerm {
		//候选者任期小于直接返回false，并且reply其他节点大任期
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}
	reply.Term = args.Term

	if rf.votedFor != -1 && rf.votedFor != args.CandidateId {
		//DPrintf("[ElectionReject+++++++]Server %d reject %d, Have voter for %d",rf.me,args.CandidateId,rf.votedFor)
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	}
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {

		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		fmt.Printf("rf.me:%d----rf.currentTerm: %d-----rf..votedFor: %d\n", rf.me, rf.currentTerm, rf.votedFor)
		rf.lastResetElectionTime = time.Now()
	} else {
		reply.VoteGranted = false
	}
}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}
