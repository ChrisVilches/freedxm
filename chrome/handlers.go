package chrome

import (
	"strings"

	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/gorilla/websocket"
)

var activeTabs = make(map[string]string)

// NOTE: This event is triggered even for targets that
// were skipped during attachment.
func handleTargetDestroyed(response cdpResponse) {
	type T struct {
		TargetID string `json:"targetId"`
	}

	params, err := util.Unmarshal[T](response.Params)
	if err != nil {
		return
	}

	delete(activeTabs, params.TargetID)
}

func handleTargetCreated(conn *websocket.Conn, response cdpResponse) {
	type T struct {
		TargetInfo targetInfo `json:"targetInfo"`
	}

	params, err := util.Unmarshal[T](response.Params)
	if err != nil {
		return
	}

	if params.TargetInfo.Type == "page" {
		executeCmd("Target.attachToTarget", "", conn, map[string]any{
			"targetId": params.TargetInfo.TargetID,
			"flatten":  true,
		})
	}
}

func handleAttachedToTarget(
	conn *websocket.Conn,
	response cdpResponse,
	matcher *patterns.Matcher,
) {
	type T struct {
		SessionID  string     `json:"sessionId"`
		TargetInfo targetInfo `json:"targetInfo"`
	}
	params, err := util.Unmarshal[T](response.Params)
	if err != nil {
		return
	}

	activeTabs[params.TargetInfo.TargetID] = params.SessionID
	blockTargetIfMatches(params.TargetInfo, params.SessionID, conn, matcher)
}

func shouldSkip(targetInfo targetInfo) bool {
	if len(targetInfo.URL) == 0 {
		return true
	}

	if targetInfo.Type != "page" {
		return true
	}

	return strings.HasPrefix(targetInfo.URL, chromeExtensionPrefix)
}

func blockTargetIfMatches(
	targetInfo targetInfo,
	sessionID string,
	conn *websocket.Conn,
	matcher *patterns.Matcher,
) {
	if shouldSkip(targetInfo) {
		return
	}

	if patternMatch := matcher.MatchesAny(targetInfo.URL); patternMatch != nil {
		executeCmd("Page.navigate",
			sessionID,
			conn,
			map[string]any{"url": redirectURL},
		)
	}
}

func handleTargetInfoChanged(
	conn *websocket.Conn,
	response cdpResponse,
	matcher *patterns.Matcher,
) {
	type T struct {
		TargetInfo targetInfo `json:"targetInfo"`
	}

	params, err := util.Unmarshal[T](response.Params)
	if err != nil {
		return
	}

	if sessionID, exists := activeTabs[params.TargetInfo.TargetID]; exists {
		blockTargetIfMatches(params.TargetInfo, sessionID, conn, matcher)
	}
}
