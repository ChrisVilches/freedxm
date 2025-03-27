package chrome

import "encoding/json"

type cdpRequest struct {
	ID        int32  `json:"id"`
	Method    string `json:"method"`
	Params    any    `json:"params,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
}

type cdpResponse struct {
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

type targetInfo struct {
	TargetID string `json:"targetId"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

type targetSession struct {
	SessionID  string     `json:"sessionId"`
	TargetInfo targetInfo `json:"targetInfo"`
}
