package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ChrisVilches/freedxm/chrome"
	"github.com/ChrisVilches/freedxm/config"
	"github.com/ChrisVilches/freedxm/fileutil"
	"github.com/ChrisVilches/freedxm/killer"
	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
	"github.com/ChrisVilches/freedxm/session"
)

var procBlackList patterns.Matcher
var domainsMatcher patterns.Matcher

func handleProcess(proc process.Process) {
	if !domainsMatcher.IsEmpty() {
		if proc.Name == "chrome" {
			go chrome.IdempotentStartChromeManager(&domainsMatcher)
		}
		if proc.Name == "firefox" {
			fmt.Println("handle firefox (dummy logic)")
		}
	}

	// TODO: Maybe we need a more strict matching mechanism here, because
	// some keywords may try to kill a lot of processes that we don't want to kill.

	// TODO: again, using an inconsistent term "black list"
	if procBlackList.MatchesAny(proc.Name) != nil {
		killer.KillAll(proc.Name)
	}
}

func onePoll() error {
	// TODO: These process names are unique (but if there are multiple, the PIDs are simplified to
	// only one instance of the process, so information is lost).
	// Somehow make it more clear what this function does.
	result, err := process.GetProcessList()

	if err != nil {
		return err
	}

	for _, proc := range result {
		handleProcess(proc)
	}

	return nil
}

func shouldPoll() bool {
	return !domainsMatcher.IsEmpty() || !procBlackList.IsEmpty()
}

var cond sync.Cond

func pollingProcess() {
	var mu sync.Mutex
	cond = *sync.NewCond(&mu)

	for {
		// TODO: Audit this cond Lock/Unlock mechanism, and
		// it's thread safe or not.
		cond.L.Lock()
		if !shouldPoll() {
			fmt.Println("Stopped")
			cond.Wait()
			fmt.Println("Continued from cond.Wait()")
		}

		fmt.Println()
		// TODO: Test this behaves as a critical zone because we can't add new elements while it's polling.
		// ^^^ I think it's fine. The matcher setters and getters are guarded, and sometimes there might be a slight
		// delay between the domain settings and the actual polling results but it will be corrected in one tick.
		err := onePoll()

		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		// TODO: Why did i put this unlock here? does
		// exiting .Wait Lock it?
		cond.L.Unlock()
		time.Sleep(1 * time.Second)
	}
}

// The mechanism to update matchers is designed to be concurrent but not fully synchronized.
// Any synchronization issues will be resolved in the next polling cycle. The most critical
// operations, such as setting or traversing matchers, are protected by mutexes to ensure
// thread safety.

// TODO: i use the terms "proc" and "programs". Choose just one.
func setMatchers(proc, domains []string) {
	procBlackList.Set(proc)
	domainsMatcher.Set(domains)
	fmt.Println(proc, domains)
	cond.Signal()
}

var pipePath = "/tmp/freedxm"
var currSessions session.CurrentSessions

func listenNewSessions() {
	var muSessions sync.Mutex

	sessionsCh, errCh := fileutil.ReadFromPipe[session.Session](pipePath)

	for {
		select {
		case s := <-sessionsCh:
			muSessions.Lock()
			sessionID := currSessions.Add(s)
			merged := currSessions.MergeLists()
			setMatchers(merged.Programs, merged.Domains)
			muSessions.Unlock()

			time.AfterFunc(time.Duration(s.TimeSeconds)*time.Second, func() {
				muSessions.Lock()
				currSessions.Remove(sessionID)
				merged := currSessions.MergeLists()
				setMatchers(merged.Programs, merged.Domains)
				muSessions.Unlock()
			})
		case err := <-errCh:
			fmt.Fprintln(os.Stderr, "Error reading from pipe:", err)
			return
		}
	}
}

// TODO: This will be the command in the actual CLI app, but
// refine its functionality first obviously.
func handleAddSessionCommand(blockListName string) {
	blockList, err := config.GetBlockListByName(blockListName)
	if err != nil {
		fmt.Println(err)
		e, ok := err.(*config.BlockListNotFoundError)
		if ok {
			fmt.Println("here are the available names", e.AvailableNames)
		}
		return
	}

	data := session.Session{
		TimeSeconds: 10,
		Programs:    blockList.Programs,
		Domains:     blockList.Domains,
	}
	err = fileutil.WriteToPipe(pipePath, data)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	blockListName, present := os.LookupEnv("A")
	if present {
		handleAddSessionCommand(blockListName)
		return
	}

	currSessions = session.NewCurrentSessions()

	fileutil.ResetFile(pipePath)

	go pollingProcess()
	go listenNewSessions()

	select {}
}
