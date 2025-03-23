package session

// TODO: This struct is repeated in another place, but without the Name field...
type BlockList struct {
	Domains  []string
	Programs []string
}

type Session struct {
	TimeSeconds int      `json:"time_seconds"`
	Domains     []string `json:"domains"`
	Programs    []string `json:"programs"`
}

type CurrentSessions struct {
	currID   uint32
	sessions map[uint32]Session
}

func NewCurrentSessions() CurrentSessions {
	return CurrentSessions{
		currID:   0,
		sessions: make(map[uint32]Session),
	}
}

// TODO: Make it thread safe. But I think the main usage is from the pipe.
// which is a sequential stream.

func (c *CurrentSessions) Add(session Session) uint32 {
	c.currID++
	c.sessions[c.currID] = session
	return c.currID
}

func (c *CurrentSessions) Remove(id uint32) {
	delete(c.sessions, id)
}

func setToSlice(set map[string]struct{}) []string {
	res := []string{}
	for key := range set {
		res = append(res, key)
	}
	return res
}

func (c CurrentSessions) MergeLists() BlockList {
	uniqDomains := make(map[string]struct{})
	uniqPrograms := make(map[string]struct{})

	for _, session := range c.sessions {
		for _, dom := range session.Domains {
			uniqDomains[dom] = struct{}{}
		}
		for _, prog := range session.Programs {
			uniqPrograms[prog] = struct{}{}
		}
	}

	return BlockList{
		Domains:  setToSlice(uniqDomains),
		Programs: setToSlice(uniqPrograms),
	}
}
