package chrome

import (
	"strings"

	"github.com/ChrisVilches/freedxm/notifier"
	"github.com/ChrisVilches/freedxm/patterns"
	"github.com/ChrisVilches/freedxm/util"
	"github.com/gorilla/websocket"
)

var activeTabs = make(map[string]targetSession)

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
	params, err := util.Unmarshal[targetSession](response.Params)
	if err != nil {
		return
	}

	activeTabs[params.TargetInfo.TargetID] = params
	blockTargetIfMatches(params, conn, matcher)
}

func shouldSkip(targetInfo targetInfo) bool {
	if len(targetInfo.URL) == 0 {
		return true
	}

	if targetInfo.Type != "page" {
		return true
	}

	for _, prefix := range skipPrefix {
		if strings.HasPrefix(targetInfo.URL, prefix) {
			return true
		}
	}

	return false
}

func blockTargetIfMatches(
	targetSession targetSession,
	conn *websocket.Conn,
	matcher *patterns.Matcher,
) {
	if shouldSkip(targetSession.TargetInfo) {
		return
	}

	patternMatch := matcher.MatchesAny(targetSession.TargetInfo.URL)

	if patternMatch != nil {
		executeCmd("Page.navigate",
			targetSession.SessionID,
			conn,
			map[string]any{"url": getRedirectURL(targetSession.TargetInfo.URL)},
		)
		notifier.NotifyWarn("Blocked Website", targetSession.TargetInfo.URL)
	}
}

func blockTargetIfMatchesAll(conn *websocket.Conn, matcher *patterns.Matcher) {
	for _, targetSession := range activeTabs {
		blockTargetIfMatches(targetSession, conn, matcher)
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

	if targetSession, exists := activeTabs[params.TargetInfo.TargetID]; exists {
		targetSession.TargetInfo = params.TargetInfo
		activeTabs[params.TargetInfo.TargetID] = targetSession

		blockTargetIfMatches(targetSession, conn, matcher)
	}
}
