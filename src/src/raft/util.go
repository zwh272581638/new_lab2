package raft

import (
	"log"
	"math/rand"
	"time"
)

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func getRandElectTimeout(server int64) int {
	rand.Seed(time.Now().Unix() + server)
	return rand.Intn(ElectionTimeoutMax-ElectionTimeoutMin) + ElectionTimeoutMin
}
func (l *Log) at(idx int) *Entry {
	return &l.Entries[idx]
}

func makeEmptyLog() Log {
	log := Log{
		Entries: make([]Entry, 0),
	}
	return log
}

func (l *Log) truncate(idx int) {
	l.Entries = l.Entries[:idx]
}

func (l *Log) slice(idx int) []Entry {
	return l.Entries[idx:]
}

func (l *Log) len() int {
	return len(l.Entries)
}

func (l *Log) lastLog() *Entry {
	return l.at(l.len() - 1)
}
func (rf *Raft) lastLogIndex() int {
	return len(rf.log.Entries) - 1
}

func (rf *Raft) getLastIndex() int {
	return len(rf.log.Entries) - 1
}

func (rf *Raft) getLastTerm() int {
	return rf.log.Entries[len(rf.log.Entries)-1].Term
}
func (rf *Raft) getLogTermWithIndex(globalIndex int) int {
	return rf.log.Entries[globalIndex].Term
}

func (rf *Raft) getLogWithIndex(globalIndex int) Entry {

	return rf.log.Entries[globalIndex]
}
func (rf *Raft) UpToDate(index int, term int) bool {
	//lastEntry := rf.log[len(rf.log)-1]
	lastIndex := rf.getLastIndex()
	lastTerm := rf.getLastTerm()
	return term > lastTerm || (term == lastTerm && index >= lastIndex)
}

//func (rf *Raft) getPrevLogInfo(server int) (int,int){
//	newEntryBeginIndex := rf.nextIndex[server]-1
//	lastIndex := rf.getLastIndex()
//	if newEntryBeginIndex == lastIndex+1 {
//		newEntryBeginIndex = lastIndex
//	}
//	return newEntryBeginIndex ,rf.getLogTermWithIndex(newEntryBeginIndex)
//}
