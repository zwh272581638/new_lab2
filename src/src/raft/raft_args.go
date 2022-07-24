package raft

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}
type Log struct {
	Entries []Entry
}
type Entry struct {
	Term    int         // leader收到日志条目时的任期
	Command interface{} // 状态机的command
}

type AppendEntriesArgs struct {
	Term         int     // leader的任期
	LeaderId     int     // leader自身的ID
	PrevLogIndex int     // 新的日志目前的索引值（预计要从哪里追加）
	PrevLogTerm  int     // 新的日志目前的任期号
	Entries      []Entry // 预计存储的日志（为空时就是心跳连接）
	LeaderCommit int     // leader的commit index指的是最后一个被大多数机器都复制的日志Index
}

type AppendEntriesReply struct {
	Term          int  // leader的term可能是过时的，此时收到的Term用于更新他自己
	Success       bool //	如果follower与Args中的PreLogIndex/PreLogTerm都匹配才会接过去新的日志（追加），不匹配直接返回false
	ConflictIndex int
	ConflictTerm  int
}
