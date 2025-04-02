package queue

import (
	"sync"

	"github.com/gkettani/bobber-the-swe/internal/models"
)

// JobQueue is a thread-safe in-memory queue for job listings
type JobQueue struct {
	items []*models.JobListing
	mutex sync.Mutex
}

// NewJobQueue creates a new in-memory job queue
func NewJobQueue() *JobQueue {
	return &JobQueue{
		items: make([]*models.JobListing, 0),
	}
}

// Enqueue adds a job listing to the queue
func (q *JobQueue) Enqueue(job *models.JobListing) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, job)
}

// Dequeue removes and returns the next job listing from the queue
// Returns nil if the queue is empty
func (q *JobQueue) Dequeue() *models.JobListing {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	job := q.items[0]
	q.items = q.items[1:]
	return job
}

// Size returns the current number of items in the queue
func (q *JobQueue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}

// IsEmpty returns true if the queue is empty
func (q *JobQueue) IsEmpty() bool {
	return q.Size() == 0
}
