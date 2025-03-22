package main

import (
	"log"
	"time"

	"github.com/ChrisVilches/freedxm/chrome"
	"github.com/ChrisVilches/freedxm/killer"
	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
)

var procBlackList = map[string]struct{}{
	"nemo":    {},
	"firefox": {},
}

var domainsMatcher = patterns.NewMatcher([]string{"aRomA-%%%.%", "a-%%%%%%%de"})

func handleChrome() {
	if !domainsMatcher.IsEmpty() {
		go chrome.EnsureChromeManager(domainsMatcher)
	}
}

func handleFirefox() {

}

func onePoll() error {
	result, err := process.GetProcessList()

	if err != nil {
		return err
	}

	for _, proc := range result {
		if proc.Name == "chrome" {
			handleChrome()
			handleFirefox()
		}

		if _, exists := procBlackList[proc.Name]; exists {
			killer.KillAll(proc.Name)
		}
	}

	return nil
}

func main() {
	for {
		err := onePoll()

		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		time.Sleep(1 * time.Second)
	}
}
