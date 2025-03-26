package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/ChrisVilches/freedxm/chrome"
	"github.com/ChrisVilches/freedxm/model"
	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
	"github.com/ChrisVilches/freedxm/rpc/implementation"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/status"
)

var processMatcher patterns.Matcher
var domainsMatcher patterns.Matcher
var defaultPort = 8687
var secondsInMinute int = 60
var doPoll = make(chan struct{}, 1)
var activePoll atomic.Bool
var chromeMonitor = util.NewIdempotentRunner(func(ctx context.Context) {
	chrome.MonitorChrome(ctx, &domainsMatcher)
})

func handleProcesses(ctx context.Context, procName string) {
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
				handleProcesses(ctx, proc)
			}

			go sleepPoll()
		}
	}
}

// The mechanism to update matchers is designed to be concurrent but not fully
// synchronized. Any synchronization issues will be resolved in the next polling
// cycle. The most critical operations, such as setting or traversing matchers,
// are protected by mutexes to ensure thread safety.

func setMatchers(processes, domains []string) {
	processMatcher.Set(processes)
	domainsMatcher.Set(domains)
	log.Println("processes:", processes, "domains:", domains)

	if activePoll.CompareAndSwap(false, true) {
		doPoll <- struct{}{}
	}
}

func serve(ctx context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	currSessions := model.NewCurrentSessions()
	errCh := make(chan error, 2)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() { errCh <- pollingProcess(ctx) }()
	go func() { errCh <- rpc.GRPCServerStart(ctx, port, &currSessions) }()

	for {
		select {
		case err := <-errCh:
			return fmt.Errorf("unexpected termination: %v", err)
		case merged, ok := <-currSessions.MergedCh:
			if !ok {
				return fmt.Errorf("MergedCh closed unexpectedly")
			}
			setMatchers(merged.Processes, merged.Domains)
		}
	}
}

func addSession(_ context.Context, cmd *cli.Command) error {
	port := cmd.Int("port")
	seconds := int(cmd.Int("minutes")) * secondsInMinute
	blockListNames := cmd.StringSlice("block-lists")
	return rpc.CreateSession(int(port), seconds, blockListNames)
}

func listSessions(_ context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	sessions, err := rpc.ListSessions(port)
	if err != nil {
		return err
	}
	log.Println(sessions)
	return nil
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Start the server",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   int64(defaultPort),
						Usage:   "Port number to listen on",
					},
				},
				Action: serve,
			},
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "Creates new blocking session",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "minutes",
						Aliases:  []string{"m"},
						Required: true,
						Usage:    "Number of minutes",
					},
					&cli.StringSliceFlag{
						Name:     "block-lists",
						Aliases:  []string{"b"},
						Required: true,
						Usage:    "Block lists to use",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   int64(defaultPort),
						Usage:   "Port where the service is running on",
					},
				},
				Action: addSession,
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List active sessions",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   int64(defaultPort),
						Usage:   "Port where the service is running on",
					},
				},
				Action: listSessions,
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)

	if err == nil {
		return
	}

	if grpcErr, ok := status.FromError(err); ok {
		// TODO: Nice, but make it a tad prettier
		log.Fatalf("gRPC Error: %s", grpcErr.Message())
	}

	log.Fatal(err)
}
