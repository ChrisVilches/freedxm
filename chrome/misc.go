package chrome

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ChrisVilches/freedxm/util"
	"github.com/gorilla/websocket"
)

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

func getBrowserWebSocketURL() (string, error) {
	url := fmt.Sprintf("http://localhost:%d/json/version", getPort())
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type T struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}

	result, err := util.Unmarshal[T](body)
	return result.WebSocketDebuggerURL, err
}

func executeCmd(
	method, sessionID string,
	conn *websocket.Conn,
	params map[string]any,
) {
	cmd := cdpRequest{
		ID:        commandID.Add(1),
		Method:    method,
		Params:    params,
		SessionID: sessionID,
	}
	if err := conn.WriteJSON(cmd); err != nil {
		log.Println("command failed:", err)
	}
}
