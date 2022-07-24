package raft

import (
	"time"
)

func (rf *Raft) leaderSendEntries() {
	// send to every server to replicate logs to them
	for index := range rf.peers {
		if index == rf.me {
			continue
		}
		// parallel replicate logs to sever
		go func(server int) {
			rf.mu.Lock()
			if rf.state != Leader {
				rf.mu.Unlock()
				return
			}
			var entriesNeeded []Entry
			//fmt.Printf("server:%d,rf.nextIndex[server]:%d\n", server, rf.nextIndex[server])
			if len(rf.log.Entries) > rf.nextIndex[server] {
				entriesNeeded = append(entriesNeeded, rf.log.Entries[rf.nextIndex[server]:]...)
			}
			nextIndex := rf.nextIndex[server]
			if nextIndex > len(rf.log.Entries) {
				nextIndex = len(rf.log.Entries)
			}
			if nextIndex <= 0 {
				nextIndex = 1
			}
			prevLogIndex := nextIndex - 1
			aeArgs := AppendEntriesArgs{
				rf.currentTerm,
				rf.me,
				prevLogIndex,
				rf.log.Entries[prevLogIndex].Term,
				entriesNeeded,
				rf.commitIndex,
			}

			aeReply := AppendEntriesReply{}
			//aeArgs.Entries = rf.logs[rf.nextIndex[server]:]
			rf.mu.Unlock()

			re := rf.sendAppendEntries(server, &aeArgs, &aeReply)

			if re == true {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if aeReply.Term > rf.currentTerm {
					rf.becomeFollower(aeReply.Term)
					//rf.lastResetElectionTime = time.Now()
					return
				}
				if rf.state == Leader {

					if aeReply.Success {
						rf.matchIndex[server] = aeArgs.PrevLogIndex + len(aeArgs.Entries)
						rf.nextIndex[server] = rf.matchIndex[server] + 1
						for n := len(rf.log.Entries) - 1; n > rf.commitIndex; n-- {
							if rf.log.Entries[n].Term != rf.currentTerm {
								continue
							}
							// the entry has replicated to leader itself
							count := 1
							for i := range rf.peers {
								if i != rf.me && rf.matchIndex[i] >= n {
									count++
									if count > len(rf.peers)/2 {
										rf.commitIndex = n
										rf.apply()
										return
									}
								}
							}
						}
					} else {
						// the accelerated log backtracking optimization
						if aeReply.ConflictTerm == -1 {
							rf.nextIndex[server] = aeReply.ConflictIndex
							return
						}
						for i := len(rf.log.Entries) - 1; i > 0; i-- {
							if rf.log.Entries[i].Term < aeReply.ConflictTerm {
								break
							}
							if rf.log.Entries[i].Term == aeReply.ConflictTerm {
								rf.nextIndex[server] = i + 1
								return
							}
						}
						rf.nextIndex[server] = aeReply.ConflictIndex
					}
				}
			}

		}(index)

	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// rule 1
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}
	if args.Term > rf.currentTerm {
		rf.becomeFollower(args.Term)
	}
	if rf.state != Follower {
		//rf.becomeFollower(args.Term)
		rf.state = Follower
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.persist()
	}

	rf.lastResetElectionTime = time.Now()
	rf.currentTerm = args.Term
	reply.Term = args.Term
	reply.Success = false
	reply.ConflictIndex = -1
	reply.ConflictTerm = -1
	rf.persist()

	// append entries rpc 2
	if len(rf.log.Entries) <= args.PrevLogIndex {
		reply.ConflictIndex = len(rf.log.Entries)
		return
	}
	if rf.log.Entries[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.ConflictTerm = rf.log.Entries[args.PrevLogIndex].Term
		for i := args.PrevLogIndex - 1; i >= 0; i-- {
			if rf.log.Entries[i].Term != reply.ConflictTerm {
				reply.ConflictIndex = i + 1
				break
			}
		}
		return
	}
	reply.Success = true
	//fmt.Printf("%d收到心跳,当前任期是%d\n", rf.me, rf.currentTerm)

	// use args.Entries to update this peer's logs
	for i, entry := range args.Entries {
		entryIndex := i + args.PrevLogIndex + 1
		// #3, conflict occurs, truncate peer's logs
		if entryIndex < len(rf.log.Entries) && rf.log.Entries[entryIndex].Term != entry.Term {
			rf.log.Entries = rf.log.Entries[:entryIndex]
			rf.persist()
		}
		// #4, append new entries (if any)
		if entryIndex >= len(rf.log.Entries) {
			rf.log.Entries = append(rf.log.Entries, args.Entries[i:]...)
			rf.persist()
			break
		}
	}

	//rule 5
	if args.LeaderCommit > rf.commitIndex {
		//rf.updateCommitIndex(FOLLOWER, args.LeaderCommit)
		lastNewIndex := rf.getLastIndex()
		if args.LeaderCommit >= lastNewIndex {
			rf.commitIndex = lastNewIndex
		} else {
			rf.commitIndex = args.LeaderCommit
		}
		rf.apply()
	}
	return
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}
