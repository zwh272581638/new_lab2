package raft

import (
	"fmt"
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"
	//	"6.824/labgob"
	"6.824/labrpc"
)

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	state       State
	currentTerm int
	votedFor    int
	logs        []LogEntry
	getVoteNum  int

	commitIndex int // 状态机中已知的被提交的日志条目的索引值(初始化为0，持续递增）
	lastApplied int // 最后一个被追加到状态机日志的索引值

	lastResetElectionTime time.Time
	applyCh               chan ApplyMsg
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == Leader
}

func (rf *Raft) electionTimeoutTicker() {
	for rf.killed() == false {
		nowTime := time.Now()
		timeElectionOut := getRandElectTimeout(int64(rf.me))
		time.Sleep(time.Duration(timeElectionOut) * time.Millisecond)
		rf.mu.Lock()

		if rf.lastResetElectionTime.Before(nowTime) && rf.state != Leader {
			fmt.Printf("%d超时变为candidate,当前任期%d\n", rf.me, rf.currentTerm+1)
			rf.becomeCandidate()
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) leaderAppendEntriesTicker() {
	for rf.killed() == false {
		time.Sleep(HeartBeatTimeOut * time.Millisecond)
		rf.mu.Lock()
		if rf.state == Leader {
			rf.lastResetElectionTime = time.Now()
			rf.mu.Unlock()
			rf.leaderSendEntries()
		} else {
			rf.mu.Unlock()
		}
	}
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	// Your initialization code here (2A, 2B, 2C).
	rf.state = Follower
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.logs = make([]LogEntry, 0)
	rf.getVoteNum = 0

	rf.commitIndex = -1
	rf.lastApplied = -1
	rf.lastResetElectionTime = time.Now()
	rf.applyCh = applyCh
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.electionTimeoutTicker()
	go rf.leaderAppendEntriesTicker()

	return rf
}

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true
	// Your code here (2B).

	return index, term, isLeader
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}
