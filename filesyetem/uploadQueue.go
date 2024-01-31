package filesyetem

import (
	"sync"
	"time"
)

type UploadQueue struct {
	items map[string]time.Time
	lock  sync.Mutex
}

func NewUploadQueue() *UploadQueue {
	return &UploadQueue{
		items: make(map[string]time.Time),
	}
}

func (q *UploadQueue) Enqueue(path string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items[path] = time.Now()
}

func (q *UploadQueue) processToChannel(output chan<- string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	currentTime := time.Now()
	for path, t := range q.items {
		if currentTime.Sub(t) > 5*time.Second {
			delete(q.items, path)
			output <- path
		}
	}
}

func (q *UploadQueue) ProcessRoutine(output chan<- string) {
	for {
		q.processToChannel(output)
		time.Sleep(1 * time.Second)
	}
}
