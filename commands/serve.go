package commands

// The mechanism to update matchers is designed to be concurrent but not fully
// synchronized. Any synchronization issues will be resolved in the next polling
// cycle. The most critical operations, such as setting or traversing matchers,
// are protected by mutexes to ensure thread safety.

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/ChrisVilches/freedxm/chrome"
	"github.com/ChrisVilches/freedxm/model"
	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
	rpc "github.com/ChrisVilches/freedxm/rpc/implementation"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/urfave/cli/v3"
)

const ServeDefaultPort = 8687

var doPoll = make(chan struct{}, 1)
var activePoll atomic.Bool
var processMatcher patterns.Matcher
var domainsMatcher patterns.Matcher
var chromeCh = make(chan struct{}, 1)

var chromeMonitor = util.NewIdempotentRunner(func(ctx context.Context) {
	chrome.MonitorChrome(ctx, &domainsMatcher, chromeCh)
})

func handleProcess(ctx context.Context, procName string) {
	if !domainsMatcher.IsEmpty() {
		if procName == "chrome" {
			go chromeMonitor.Run(ctx)
		}
		if procName == "firefox" {
			log.Println("handle firefox (dummy logic)")
		}
	}

	if processMatcher.MatchesAny(procName) != nil {
		process.KillAll(procName)
	}
}

func setMatchers(processes, domains []string) {
	processMatcher.Set(processes)
	domainsMatcher.Set(domains)
	log.Println("processes:", processes, "domains:", domains)

	if activePoll.CompareAndSwap(false, true) {
		doPoll <- struct{}{}
	}
}

func listenSessionsUpdated(
	currSessions *model.CurrentSessions,
	sessionsUpdateCh <-chan struct{},
) {
	observers := []chan<- struct{}{chromeCh}
	for {
		<-sessionsUpdateCh
		res := currSessions.MergeLists()
		setMatchers(res.Processes, res.Domains)

		for _, ch := range observers {
			select {
			case ch <- struct{}{}:
			default:
				log.Println("observer skipped when notifying sessions change")
			}
		}
	}
}

func sleepPoll() {
	time.Sleep(1 * time.Second)

	if !domainsMatcher.IsEmpty() || !processMatcher.IsEmpty() {
		doPoll <- struct{}{}
	} else {
		activePoll.Store(false)
	}
}

func pollingProcess(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-doPoll:
			result, err := process.GetProcessNames()

			if err != nil {
				return err
			}

			for _, proc := range result {
				handleProcess(ctx, proc)
			}

			go sleepPoll()
		}
	}
}

func Serve(ctx context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	sessionsUpdateCh := make(chan struct{})
	currSessions := model.NewCurrentSessions()
	errCh := make(chan error, 2)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() { errCh <- pollingProcess(ctx) }()
	go func() {
		errCh <- rpc.GRPCServerStart(ctx, port, &currSessions, sessionsUpdateCh)
	}()
	go listenSessionsUpdated(&currSessions, sessionsUpdateCh)

	err := <-errCh
	return fmt.Errorf("unexpected termination: %v", err)
}
