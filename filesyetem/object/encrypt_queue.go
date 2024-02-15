package object

import (
	"fmt"
	"sync"
	"time"
)

type EncryptQueue struct {
	items map[string]time.Time
	lock  sync.Mutex
}

func (f *Service) NewEncryptQueue() *EncryptQueue {
	q := &EncryptQueue{
		items: make(map[string]time.Time),
	}
	go q.StartQueueRoutine(f.encryptChan)
	return q
}

func (q *EncryptQueue) Enqueue(id string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items[id] = time.Now()
}

func (q *EncryptQueue) processToChannel(output chan<- encryptChanItem) {
	q.lock.Lock()
	defer q.lock.Unlock()

	currentTime := time.Now()
	for id, t := range q.items {
		if currentTime.Sub(t) > 5*time.Second {
			delete(q.items, id)
			q.lock.Unlock()
			output <- encryptChanItem{id: id}
			q.lock.Lock()
			fmt.Printf("Upload Queued: %s \n", id)
		}
	}
}

func (q *EncryptQueue) StartQueueRoutine(output chan<- encryptChanItem) {
	for {
		q.processToChannel(output)
		time.Sleep(1 * time.Second)
	}
}

func (q *EncryptQueue) IsInQueue(id string) bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	_, is := q.items[id]
	return is
}
