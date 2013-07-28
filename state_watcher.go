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

func (sw *StateWatcher) OnModify(callback func(io.Reader)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-watcher.Event:
				body, err := os.Open(sw.StateFilePath)
				if err != nil {
					break
				}

				callback(body)
			case <-watcher.Error:
			}
		}
	}()

	return watcher.WatchFlags(sw.StateFilePath, fsnotify.FSN_MODIFY)
}
