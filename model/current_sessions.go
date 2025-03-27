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
	mu       sync.Mutex
}

func NewCurrentSessions() CurrentSessions {
	return CurrentSessions{
		currID:   0,
		sessions: make(map[uint32]Session),
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

func (c *CurrentSessions) Add(session Session) uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currID++
	c.sessions[c.currID] = session
	return c.currID
}

func (c *CurrentSessions) Remove(id uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, id)
}

func (c *CurrentSessions) MergeLists() MergeResult {
	c.mu.Lock()
	defer c.mu.Unlock()

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
