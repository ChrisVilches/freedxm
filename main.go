package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ChrisVilches/freedxm/chrome"
	"github.com/ChrisVilches/freedxm/http"
	"github.com/ChrisVilches/freedxm/model"
	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/urfave/cli/v3"
)

var processMatcher patterns.Matcher
var domainsMatcher patterns.Matcher
var defaultPort = 8687
var secondsInMinute int = 60

func handleProcesses(procName string) {
	if !domainsMatcher.IsEmpty() {
		if procName == "chrome" {
			go chrome.IdempotentStartChromeManager(&domainsMatcher)
		}
		if procName == "firefox" {
			log.Println("handle firefox (dummy logic)")
		}
	}

	if processMatcher.MatchesAny(procName) != nil {
		process.KillAll(procName)
	}
}

var pollCondMtx sync.Mutex
var pollCondVar util.CondVar

func pollingProcess() {
	pollCondVar = *util.NewCondVar(&pollCondMtx)

	for {
		pollCondVar.WaitUntil(func() bool {
			return !domainsMatcher.IsEmpty() || !processMatcher.IsEmpty()
		})

		result, err := process.GetProcessNames()

		if err != nil {
			log.Print(err)
			return
		}

		for _, proc := range result {
			handleProcesses(proc)
		}

		time.Sleep(1 * time.Second)
	}
}

// The mechanism to update matchers is designed to be concurrent but not fully
// synchronized. Any synchronization issues will be resolved in the next polling
// cycle. The most critical operations, such as setting or traversing matchers,
// are protected by mutexes to ensure thread safety.

func setMatchers(processes, domains []string) {
	pollCondMtx.Lock()
	defer pollCondMtx.Unlock()

	processMatcher.Set(processes)
	domainsMatcher.Set(domains)
	log.Println("processes:", processes, "domains:", domains)
	pollCondVar.Signal()
}

func serve(_ context.Context, cmd *cli.Command) error {
	port := int(cmd.Int("port"))
	currSessions := model.NewCurrentSessions()

	go pollingProcess()
	go http.StartHTTPServer(port, &currSessions)

	for merged := range currSessions.MergedCh {
		setMatchers(merged.Processes, merged.Domains)
	}

	return fmt.Errorf(
		"unexpected termination: the service process should run indefinitely",
	)
}

func addSession(_ context.Context, cmd *cli.Command) error {
	port := cmd.Int("port")
	seconds := int(cmd.Int("minutes")) * secondsInMinute
	blockListNames := cmd.StringSlice("block-lists")
	return http.CreateSession(int(port), seconds, blockListNames)
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
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
