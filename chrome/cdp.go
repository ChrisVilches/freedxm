package chrome

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/gorilla/websocket"
)

const (
	defaultPort          = 9222
	wsChSize             = 10
	blockPagePath        = "/tmp/freedxm-block.html"
	blockPagePermissions = 0644
)

//go:embed 1.html
var blockPageEmbeddedHTML string

var skipPrefix = []string{
	"chrome-extension://",
	"file://",
}

var commandID atomic.Int32

func createConnection() (*websocket.Conn, error) {
	wsURL, err := getBrowserWebSocketURL()
	if err != nil {
		return nil, err
	}
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func readResponse(conn *websocket.Conn) (cdpResponse, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return cdpResponse{}, err
	}

	return util.Unmarshal[cdpResponse](message)
}

func handleNoDebugger() {
	log.Printf("No debugger, must kill Chrome")
	process.KillAll("chrome")
}

func handleResponse(
	response cdpResponse,
	conn *websocket.Conn,
	matcher *patterns.Matcher,
) {
	switch response.Method {
	case "Target.targetDestroyed":
		handleTargetDestroyed(response)

	case "Target.targetCreated":
		handleTargetCreated(conn, response)

	case "Target.attachedToTarget":
		handleAttachedToTarget(conn, response, matcher)

	case "Target.targetInfoChanged":
		handleTargetInfoChanged(conn, response, matcher)
	}
}

func getMsgsFromConnection(conn *websocket.Conn) <-chan cdpResponse {
	ch := make(chan cdpResponse, wsChSize)
	go func() {
		for {
			response, err := readResponse(conn)
			if err != nil {
				close(ch)
				return
			}
			ch <- response
		}
	}()
	return ch
}

func getRedirectURL(blockedURL string) string {
	encodedURL := url.QueryEscape(blockedURL)
	return fmt.Sprintf("file://%s?url=%s", blockPagePath, encodedURL)
}

func createBlockHTMLPage() {
	err := os.WriteFile(
		blockPagePath,
		[]byte(blockPageEmbeddedHTML),
		blockPagePermissions,
	)
	if err != nil {
		log.Println("Failed to write embedded HTML to file:", err)
	}
}

// When a session starts, it blocks existing tabs using the target
// attaching mechanism. For additional sessions, an "update" event
// triggers, attempting to block all tabs that match the new domain blocklist.
func MonitorChrome(ctx context.Context, matcher *patterns.Matcher, updateCh <-chan struct{}) {
	createBlockHTMLPage()

	conn, err := createConnection()
	if err != nil {
		handleNoDebugger()
		return
	}

	commandID.Store(1)
	defer conn.Close()

	executeCmd("Target.setDiscoverTargets",
		"",
		conn,
		map[string]any{"discover": true},
	)

	log.Println("started monitoring chrome")

	responseCh := getMsgsFromConnection(conn)

	for {
		select {
		case <-ctx.Done():
			log.Println("finished chrome monitoring (context done)")
			return
		case <-updateCh:
			if matcher.IsEmpty() {
				log.Println("finished chrome monitoring (no blocked domains)")
				return
			}
			blockTargetIfMatchesAll(conn, matcher)
		case response, ok := <-responseCh:
			if !ok {
				log.Println("websocket channel closed")
				return
			}

			handleResponse(response, conn, matcher)
		}
	}
}
