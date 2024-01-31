package fuse

import (
	"fmt"
	"sync"
	"time"
)

type UploadQueue struct {
	items map[string]time.Time
	lock  sync.Mutex
}

func (q *UploadQueue) Enqueue(path string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items[path] = time.Now()
}

func (q *UploadQueue) Dequeue(path string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	delete(q.items, path)
}

func (q *UploadQueue) Process(output chan<- string) {
	for {
		q.lock.Lock()
		currentTime := time.Now()
		for key, value := range q.items {
			if currentTime.Sub(value) < 5*time.Second {
				q.Dequeue(key)
				fmt.Printf("aaaa: %s", key)
				output <- key
			}
		}
		q.lock.Unlock()
		time.Sleep(1 * time.Second)
	}
}
