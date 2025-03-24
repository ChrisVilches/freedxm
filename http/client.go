package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func CreateSession(port, timeSeconds int, blockListNames []string) error {
	client := resty.New()

	payload := newSessionPayload{
		TimeSeconds: timeSeconds,
		BlockLists:  blockListNames,
	}

	var errResp errorResponse

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetError(&errResp).
		Post(fmt.Sprintf("http://localhost:%d/session", port))

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(errResp.Error)
	}

	return nil
}
