package raft

import (
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
	log         Log
	getVoteNum  int

	nextIndex   []int //下一个待发送的日志索引，leader特有
	matchIndex  []int //已同步的日志索引，leader特有
	commitIndex int   // 状态机中已知的被提交的日志条目的索引值(初始化为0，持续递增）
	lastApplied int   // 最后一个被追加到状态机日志的索引值

	lastResetElectionTime time.Time
	applyCh               chan ApplyMsg
	applyCond             *sync.Cond // condition to trigger apply log entry
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
			//fmt.Printf("%d超时变为candidate,当前任期%d\n", rf.me, rf.currentTerm+1)
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
func (rf *Raft) committedToAppliedTicker() {
	// put the committed entry to apply on the state machine
	for rf.killed() == false {
		time.Sleep(AppliedTimeOut * time.Millisecond)
		rf.mu.Lock()

		if rf.lastApplied >= rf.commitIndex {
			rf.mu.Unlock()
			continue
		}

		Messages := make([]ApplyMsg, 0)
		for rf.lastApplied < rf.commitIndex && rf.lastApplied < rf.getLastIndex() {
			//for rf.lastApplied < rf.commitIndex {
			rf.lastApplied += 1
			Messages = append(Messages, ApplyMsg{
				CommandValid:  true,
				SnapshotValid: false,
				CommandIndex:  rf.lastApplied,
				Command:       rf.getLogWithIndex(rf.lastApplied).Command,
			})
		}
		rf.mu.Unlock()

		for _, messages := range Messages {
			rf.applyCh <- messages
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
	rf.log.Entries = append(rf.log.Entries, Entry{0, 0})
	rf.getVoteNum = 0

	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	rf.lastResetElectionTime = time.Now()
	rf.applyCh = applyCh
	rf.applyCond = sync.NewCond(&rf.mu)
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.electionTimeoutTicker()
	go rf.leaderAppendEntriesTicker()
	//go rf.committedToAppliedTicker()
	go rf.applier()
	return rf
}

//对leader节点写入log

func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := -1
	isLeader := true
	// Your code here (2B).
	if rf.killed() {
		return index, term, false
	}
	if rf.state != Leader {
		return index, term, false
	}
	isLeader = true

	term = rf.currentTerm
	// 初始化日志条目。并进行追加
	appendLog := Entry{Term: rf.currentTerm, Command: command}
	rf.log.Entries = append(rf.log.Entries, appendLog)
	index = len(rf.log.Entries) - 1

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
