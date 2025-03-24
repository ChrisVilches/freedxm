package chrome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
)

// TODO: This doesn't work for domains or query params with kanji.
// The matcher itself does work (regex is working), but the Chrome
// debugger returns the domains with encrypted texts so I can't
// compare them against the matchers.

type debuggerInfo struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func getPort() int {
	portStr, present := os.LookupEnv("CHROME_DEBUGGER_PORT")

	if !present {
		return defaultPort
	}

	port, err := strconv.Atoi(portStr)

	if err != nil {
		log.Printf(
			"Invalid Chrome debugger port '%s' (using default port %d)",
			portStr,
			defaultPort,
		)
		return defaultPort
	}

	return port
}

var (
	defaultPort           = 9222
	port                  = getPort()
	running               atomic.Bool
	chromeExtensionPrefix = "chrome-extension://"
	sleepTime             = 1000 * time.Millisecond
)

func closePage(id string) error {
	url := fmt.Sprintf("http://localhost:%d/json/close/%s", port, id)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func queryChromeDebugger(port int) ([]debuggerInfo, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/json", port))
	if err != nil {
		return nil, fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %v", err)
	}

	var data []debuggerInfo
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON: %v", err)
	}

	return data, nil
}

func shouldSkip(item debuggerInfo) bool {
	if len(item.URL) == 0 {
		return true
	}

	if item.Type != "page" {
		return true
	}

	return strings.HasPrefix(item.URL, chromeExtensionPrefix)
}

func hasDebugger() bool {
	_, err := queryChromeDebugger(port)
	return err == nil
}

func handleSite(matcher *patterns.Matcher, item debuggerInfo) {
	if patternMatch := matcher.MatchesAny(item.URL); patternMatch != nil {
		err := closePage(item.ID)
		if err == nil {
			log.Printf("closed %s (matches '%s')", item.URL, *patternMatch)
		} else {
			log.Printf("failed to close page (%v)", err)
		}
	}
}

func manage(matcher *patterns.Matcher) bool {
	data, err := queryChromeDebugger(port)
	if err != nil {
		return false
	}

	if matcher.IsEmpty() {
		return false
	}

	for _, item := range data {
		if !shouldSkip(item) {
			handleSite(matcher, item)
		}
	}
	return true
}

func IdempotentStartChromeManager(matcher *patterns.Matcher) {
	if !running.CompareAndSwap(false, true) {
		return
	}

	defer running.Store(false)
	defer log.Println("Chrome monitoring finished")

	if !hasDebugger() {
		process.KillAll("chrome")
		log.Printf("No debugger (port %d), must kill Chrome", port)
		return
	}

	log.Println("Chrome debugger detected (monitoring initiated)")

	for manage(matcher) {
		time.Sleep(sleepTime)
	}
}
