package chrome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ChrisVilches/freedxm/killer"
	"github.com/ChrisVilches/freedxm/patterns"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// TODO: there's a change this is missing some items when deserializing. I had that sensation but when
// I tested it, it obtained all elements. Test again and verify it's deserializing all elements without skipping.
// Compare it against the vanilla query: curl http://localhost:9222/json
type DebuggerInfo struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// TODO: Sadly it's not so easy to redirect the page, due to all the CSP protection.
func closePage(id string) {
	url := fmt.Sprintf("http://localhost:9222/json/close/%s", id)

	// Create a POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Print the response status
	fmt.Println("Response Status:", resp.Status)
}

// TODO: Use this. forgot how to use Sprintf
var port = 9222

var chromeExtensionPrefix = "chrome-extension://"

func getPages(port int) ([]DebuggerInfo, error) {
	// Send a GET request to localhost:9222
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/json", port))
	if err != nil {
		return nil, fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %v", err)
	}

	// Unmarshal the JSON response into a slice of DebuggerInfo structs
	var data []DebuggerInfo
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON: %v", err)
	}

	return data, nil
}

func hasDebugger() bool {
	_, err := getPages(port)
	return err == nil
}

func shouldSkipURL(url string) bool {
	return len(url) == 0 || strings.HasPrefix(url, chromeExtensionPrefix)
}

// TODO: Write a fuckign comment
func manage(matcher patterns.Matcher) bool {
	data, err := getPages(port)
	if err != nil {
		return false
	}

	for _, item := range data {
		if shouldSkipURL(item.URL) {
			continue
		}

		if patternMatch := matcher.MatchesAny(item.URL); patternMatch != nil {
			closePage(item.ID)
			fmt.Printf("Closed %s (matches %s)\n", item.URL, *patternMatch)
			break
		}
	}

	return true
}

var running atomic.Bool

// TODO: write comment
func IsRunning() bool {
	return running.Load()
}

// TODO: write comment
func EnsureChromeManager(matcher patterns.Matcher) {
	// TODO: This most likely works correctly, but do a second check.
	if !running.CompareAndSwap(false, true) {
		return
	}

	// TODO: This glitches a bit because when I open a Chrome window it ends first and then starts again.
	// Verify what's going on.
	// TODO: Also, sometimes if I close a website, the request will fail. Verify why this happens and if it's a bug or not.
	if !hasDebugger() {
		// TODO: Test this branch (can be tested easily with i3 menu)
		killer.KillAll("chrome")
		running.Store(false)
		return
	}

	fmt.Println("doing chrome. has debugger")

	for manage(matcher) {
		time.Sleep(1 * time.Second)
	}

	running.Store(false)
	fmt.Println("chrome ended")
}
