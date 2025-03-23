package chrome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ChrisVilches/freedxm/killer"
	"github.com/ChrisVilches/freedxm/patterns"
)

// TODO: there's a change this is missing some items when deserializing. I had that sensation but when
// I tested it, it obtained all elements. Test again and verify it's deserializing all elements without skipping.
// Compare it against the vanilla query: curl http://localhost:9222/json
type debuggerInfo struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

var (
	port                  = 9222
	running               atomic.Bool
	chromeExtensionPrefix = "chrome-extension://"
	sleepTime             = 1000 * time.Millisecond
)

// TODO: Sadly it's not so easy to redirect the page,
// due to all the CSP protection.
func closePage(id string) error {
	url := fmt.Sprintf("http://localhost:9222/json/close/%s", id)

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

func shouldSkipURL(url string) bool {
	return len(url) == 0 || strings.HasPrefix(url, chromeExtensionPrefix)
}

func hasDebugger() bool {
	_, err := queryChromeDebugger(port)
	return err == nil
}

func handleSite(matcher *patterns.Matcher, item debuggerInfo) {
	if patternMatch := matcher.MatchesAny(item.URL); patternMatch != nil {
		err := closePage(item.ID)
		if err == nil {
			fmt.Printf("Closed %s (matches %s)\n", item.URL, *patternMatch)
		} else {
			fmt.Fprintf(os.Stderr, "failed to close page (%v)", err)
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
		if shouldSkipURL(item.URL) {
			continue
		}

		handleSite(matcher, item)
	}
	return true
}

// TODO: See what happens when iframes are closed.
func IdempotentStartChromeManager(matcher *patterns.Matcher) {
	// TODO: This most likely works correctly, but do a second check.
	if !running.CompareAndSwap(false, true) {
		return
	}

	defer running.Store(false)
	defer fmt.Println("chrome ended")

	// TODO: This glitches a bit because when I open a Chrome
	// window it ends first and then starts again.
	// Verify what's going on.
	// TODO: Also, sometimes if I close a website, the request will fail.
	// Verify why this happens and if it's a bug or not.
	if !hasDebugger() {
		// TODO: Test this branch (can be tested easily with i3 menu)
		killer.KillAll("chrome")
		fmt.Println("no debugger, must kill chrome")
		return
	}

	fmt.Println("doing chrome. has debugger")

	for manage(matcher) {
		time.Sleep(sleepTime)
	}
}
