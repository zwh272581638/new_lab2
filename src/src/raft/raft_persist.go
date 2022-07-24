package raft

import (
	"6.824/labgob"
	"bytes"
	"fmt"
)

func (rf *Raft) persist() {
	data := rf.persistData()
	rf.persister.SaveRaftState(data)
}

func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 {
		return
	}
	BufferBuf := bytes.NewBuffer(data) //字节流
	d := labgob.NewDecoder(BufferBuf)  //解码
	//按照存储顺序读取持久化的数据，底层可能是数组？
	var restoreCurrentTerm int
	var restoreVotedFor int
	var restoreLog Log
	var lastIncludedIndex int
	var lastIncludedTerm int
	if d.Decode(&restoreCurrentTerm) != nil ||
		d.Decode(&restoreVotedFor) != nil ||
		d.Decode(&restoreLog) != nil ||
		d.Decode(&lastIncludedIndex) != nil ||
		d.Decode(&lastIncludedTerm) != nil {
		fmt.Printf("execute readPersist error!  restoreCurrentTerm == {%d}————restoreVotedFor == {%d}————resoteLog == {%v}\n", restoreCurrentTerm, restoreVotedFor, restoreLog)
	} else {
		rf.currentTerm = restoreCurrentTerm
		rf.votedFor = restoreVotedFor
		rf.log = restoreLog
		rf.lastIncludedIndex = lastIncludedIndex
		rf.lastIncludedTerm = lastIncludedTerm
	}
}

func (rf *Raft) persistData() []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	e.Encode(rf.lastIncludedIndex)
	e.Encode(rf.lastIncludedTerm)
	data := w.Bytes()
	return data
}
