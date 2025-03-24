package http

type errorResponse struct {
	Error string `json:"error"`
}

type newSessionPayload struct {
	BlockLists  []string `json:"blockLists" binding:"required"`
	TimeSeconds int      `json:"timeSeconds" binding:"required"`
}
