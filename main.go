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

func setMatchers(proc, domains []string) {
	procBlackList.Set(proc)
	domainsMatcher.Set(domains)
	cond.Signal()
}

func main() {
	go pollingProcess()

	configDir := "./conf"
	configFile := "./block-lists.toml"

	dataCh, errCh := fileutil.ListenFileToml[config.Config](configDir, configFile)

	go func() {
		for {
			select {
			case data := <-dataCh:
				// TODO: I'm reading the first blocklist only.
				// It should be all lists later, and configure more stuff
				// like which lists does a session use, etc.
				setMatchers(data.Blocklists[0].Programs, data.Blocklists[0].Domains)
				fmt.Println("Received:", data)
			case err := <-errCh:
				fmt.Println("Error:", err)
			}
		}
	}()

	select {}
}
