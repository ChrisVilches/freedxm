package chrome

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/process"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/gorilla/websocket"
)

const (
	defaultPort           = 9222
	chromeExtensionPrefix = "chrome-extension://"
	// TODO: this is not going to work on all environments.
	// Maybe one way to solve this issue
	// is to read the HTML files beforehand and then send the
	// HTML to replace the content.
	// Either that or let the user specify the path to the file,
	// but consider that the
	// distribution of the software would need to require adding
	// those files, which I
	// don't know how to do yet (e.g. configure the pacman
	// package to include those
	// files and install them, but even then, how am I
	// going to reference them?).
	// some other alternative ideas: (1) start an http server (2) dump the HTML
	// to a /tmp file and open that path in the browser.
	redirectURL = "file:///home/chris/dev/freedxm/block-pages/1.html"
)

var commandID atomic.Int32

func createConnection(ctx context.Context) (*websocket.Conn, error) {
	wsURL, err := getBrowserWebSocketURL()
	if err != nil {
		return nil, err
	}
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		if c != nil {
			c.Close()
		}
	}()

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

func MonitorChrome(ctx context.Context, matcher *patterns.Matcher) {
	conn, err := createConnection(ctx)
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
	defer log.Println("finished monitoring chrome")

	for {
		response, err := readResponse(conn)
		if err != nil {
			log.Println("could not read response (CDP websocket)")
			break
		}

		if matcher.IsEmpty() {
			log.Println("finished chrome monitoring (no blocked domains)")
			break
		}

		handleResponse(response, conn, matcher)
	}
}
