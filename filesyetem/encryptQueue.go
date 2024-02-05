package filesyetem

import (
	"fmt"
	"sync"
	"time"
)

type EncryptQueue struct {
	items map[string]time.Time
	lock  sync.Mutex

	fs *FileSystem
}

func (f *FileSystem) NewEncryptQueue() *EncryptQueue {
	return &EncryptQueue{
		items: make(map[string]time.Time),
		fs:    f,
	}
}

func (q *EncryptQueue) Enqueue(path string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items[path] = time.Now()
}

func (q *EncryptQueue) Rename(oldPath string, newPath string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if _, ex := q.items[oldPath]; ex {
		delete(q.items, oldPath)
		q.items[newPath] = time.Now()
	}
}

func (q *EncryptQueue) processToChannel(output chan<- encryptChanItem) {
	q.lock.Lock()
	defer q.lock.Unlock()

	currentTime := time.Now()
	for path, t := range q.items {
		if currentTime.Sub(t) > 5*time.Second {
			delete(q.items, path)
			id, err := q.fs.GetFileId(path)
			if err != nil {
				continue

			}
			q.lock.Unlock()
			output <- encryptChanItem{id: id}
			q.lock.Lock()
			fmt.Printf("Upload Queued: %s \n", path)
		}
	}
}

func (q *EncryptQueue) StartQueueRoutine(output chan<- encryptChanItem) {
	for {
		q.processToChannel(output)
		time.Sleep(1 * time.Second)
	}
}

func (q *EncryptQueue) IsInQueue(path string) bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	_, is := q.items[path]
	return is
}
