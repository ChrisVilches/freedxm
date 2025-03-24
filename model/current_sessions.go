package model

import (
	"sync"

	"github.com/ChrisVilches/freedxm/util"
)

type MergeResult struct {
	Domains   []string
	Processes []string
}

type CurrentSessions struct {
	currID   uint32
	sessions map[uint32]Session
	MergedCh chan MergeResult
	mu       sync.Mutex
}

func NewCurrentSessions() CurrentSessions {
	return CurrentSessions{
		currID:   0,
		sessions: make(map[uint32]Session),
		MergedCh: make(chan MergeResult),
	}
}

func (c *CurrentSessions) GetAll() []Session {
	c.mu.Lock()
	defer c.mu.Unlock()

	ret := make([]Session, 0)
	for _, s := range c.sessions {
		ret = append(ret, s)
	}
	return ret
}

func (c *CurrentSessions) notify() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.MergedCh <- c.mergeLists()
}

func (c *CurrentSessions) Add(session Session) uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currID++
	c.sessions[c.currID] = session
	go c.notify()
	return c.currID
}

func (c *CurrentSessions) Remove(id uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	go c.notify()
	delete(c.sessions, id)
}

func (c *CurrentSessions) mergeLists() MergeResult {
	uniqDomains := make(map[string]struct{})
	uniqProcesses := make(map[string]struct{})

	for _, session := range c.sessions {
		for _, blockList := range session.BlockLists {
			util.AddSliceToSet(uniqDomains, blockList.Domains)
			util.AddSliceToSet(uniqProcesses, blockList.Processes)
		}
	}

	return MergeResult{
		Domains:   util.SetToSlice(uniqDomains),
		Processes: util.SetToSlice(uniqProcesses),
	}
}
