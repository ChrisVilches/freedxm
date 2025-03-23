package fileutil

import (
	"path"

	"github.com/fsnotify/fsnotify"
)

// TODO: These channels should be "readonly" or "writeonly" depending on the case.

func listenAux[T any](dir, filename string, dataCh chan T, errCh chan error) {
	watcher, err := fsnotify.NewWatcher()
	filePath := path.Join(dir, filename)
	if err != nil {
		errCh <- err
		return
	}

	defer watcher.Close()

	err = watcher.Add(dir)

	if err != nil {
		errCh <- err
		return
	}

	// TODO: Audit this code.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) && event.Name == filePath {
				data, err := readTomlFile[T](path.Join(dir, filename))

				if err != nil {
					errCh <- err
				} else {
					dataCh <- *data
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			errCh <- err
		}
	}
}

func ListenFileToml[T any](dir, filename string) (chan T, chan error) {
	dataCh := make(chan T)
	errCh := make(chan error)

	go func() {
		data, err := readTomlFile[T](path.Join(dir, filename))

		if err != nil {
			errCh <- err
		} else {
			dataCh <- *data
		}

		listenAux(dir, filename, dataCh, errCh)
	}()

	return dataCh, errCh
}
