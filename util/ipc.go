package util

import (
	"encoding/json"
	"log"
)

func Unmarshal[T any](data []byte) (T, error) {
	var res T
	if err := json.Unmarshal(data, &res); err != nil {
		log.Println("unmarshal error:", err)
		return res, err
	}

	return res, nil
}
