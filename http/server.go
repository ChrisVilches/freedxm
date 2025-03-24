package http

// TODO: Consider moving everything to gRPC instead of HTTP if
// that makes it simpler.

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ChrisVilches/freedxm/config"
	"github.com/ChrisVilches/freedxm/model"
	"github.com/gin-gonic/gin"
)

var currSessionsRef *model.CurrentSessions

// TODO: Implement returning "time left" as well.
// Or time created would be enough maybe.
func handleSessionFetch(c *gin.Context) {
	c.JSON(http.StatusOK, currSessionsRef.GetAll())
}

func handleCreateSession(c *gin.Context) {
	var payload newSessionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	blockLists := make([]model.BlockList, 0)

	for _, b := range payload.BlockLists {
		blockList, err := config.GetBlockListByName(b)
		if err != nil {
			e, ok := err.(*config.BlockListNotFoundError)
			if ok {
				msg := fmt.Sprintf(
					"%s (available names: %v)",
					e.Error(),
					e.AvailableNames,
				)

				c.JSON(http.StatusBadRequest, errorResponse{Error: msg})
			} else {
				c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
			}
			return
		}

		blockLists = append(blockLists, *blockList)
	}

	sessionID := currSessionsRef.Add(model.Session{
		TimeSeconds: payload.TimeSeconds,
		BlockLists:  blockLists,
	})

	log.Println("Session started")

	time.AfterFunc(time.Duration(payload.TimeSeconds)*time.Second, func() {
		log.Println("Session finished")
		currSessionsRef.Remove(sessionID)
	})

	c.JSON(http.StatusOK, nil)
}

func StartHTTPServer(port int, currSessions *model.CurrentSessions) {
	currSessionsRef = currSessions

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.GET("/sessions", handleSessionFetch)
	r.POST("/session", handleCreateSession)
	r.Run(fmt.Sprintf(":%d", port))
}
