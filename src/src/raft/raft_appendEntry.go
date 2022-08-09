package raft

import (
	"time"
)

func (rf *Raft) leaderSendEntries() {
	// send to every server to replicate logs to them
	for index := range rf.peers {
		if index == rf.me {
			//rf.lastResetElectionTime = time.Now()
			continue
		}
		// parallel replicate logs to sever
		go func(server int) {
			rf.mu.Lock()
			//if rf.state!=Leader{
			//	rf.mu.Unlock()
			//	return
			//}
			aeArgs := AppendEntriesArgs{}
			aeArgs = AppendEntriesArgs{
				rf.currentTerm,
				rf.me,
				0,
				0,
				[]LogEntry{},
				rf.commitIndex,
			}
			aeReply := AppendEntriesReply{}
			rf.mu.Unlock()

			re := rf.sendAppendEntries(server, &aeArgs, &aeReply)
			if re == true {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if rf.state != Leader {
					return
				}
				if aeReply.Term > rf.currentTerm {
					rf.becomeFollower(aeReply.Term)
					rf.lastResetElectionTime = time.Now()
					return
				}
			}

		}(index)

	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term > rf.currentTerm {
		rf.becomeFollower(args.Term)
		return
	}
	// rule 1
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}
	rf.lastResetElectionTime = time.Now()

}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}
