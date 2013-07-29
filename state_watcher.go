package basil

import (
	"github.com/howeyc/fsnotify"
	"io"
	"os"
)

type StateWatcher struct {
	StateFilePath string
}

func NewStateWatcher(filePath string) *StateWatcher {
	return &StateWatcher{
		StateFilePath: filePath,
	}
}

func (sw *StateWatcher) OnStateChange(callback func(io.Reader)) error {
	go sw.handleUpdate(callback)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-watcher.Event:
				err := sw.handleUpdate(callback)
				if err != nil {
					break
				}
			case <-watcher.Error:
			}
		}
	}()

	return watcher.WatchFlags(sw.StateFilePath, fsnotify.FSN_MODIFY)
}

func (sw *StateWatcher) handleUpdate(callback func(io.Reader)) error {
	body, err := os.Open(sw.StateFilePath)
	if err != nil {
		return err
	}

	callback(body)

	return nil
}
