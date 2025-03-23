package fileutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"
)

var permissions uint32 = 0666

func ResetFile(path string) error {
	os.Remove(path)
	return syscall.Mkfifo(path, permissions)
}

// TODO: Hangs if the pipe reading process isn't running.
// Maybe simply fix this by checking first if the main process is running.
// (do that in the CLI app, not here).
func WriteToPipe[T any](path string, data T) error {
	if err := isPipePresent(path); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func isPipePresent(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pipe does not exist at path: %s", path)
		}
		return err
	}
	if info.Mode()&os.ModeNamedPipe == 0 {
		return fmt.Errorf("path is not a named pipe: %s", path)
	}
	return nil
}

func ReadFromPipe[T any](path string) (<-chan T, <-chan error) {
	dataCh := make(chan T)
	errCh := make(chan error)

	go func() {
		for {
			// Open the pipe (blocks until data is available)
			file, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
			if err != nil {
				errCh <- err
			}

			decoder := json.NewDecoder(file)

			for {
				var data T
				err := decoder.Decode(&data)
				if err == io.EOF {
					// TODO: I think I should close the file here.
					// I should audit this whole function.
					break
				} else if err != nil {
					file.Close()
					errCh <- err
				}

				dataCh <- data
			}

			file.Close()
		}
	}()

	return dataCh, errCh
}
