package queue

import (
	"sync"

	"github.com/gkettani/bobber-the-swe/internal/metrics"
	"github.com/gkettani/bobber-the-swe/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

type JobQueue struct {
	items          []*models.JobReference
	mutex          sync.Mutex
	queueSizeGauge prometheus.Gauge
}

func NewJobQueue() *JobQueue {
	queueSizeGauge := metrics.GetManager().CreateGauge("job_reference_queue_size", "The size of the job reference queue")
	return &JobQueue{
		items:          make([]*models.JobReference, 0),
		queueSizeGauge: queueSizeGauge,
	}
}

func (q *JobQueue) Enqueue(job *models.JobReference) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, job)
	q.queueSizeGauge.Set(float64(len(q.items)))
}

func (q *JobQueue) Dequeue() *models.JobReference {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	job := q.items[0]
	q.items = q.items[1:]
	q.queueSizeGauge.Set(float64(len(q.items)))
	return job
}

func (q *JobQueue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}

func (q *JobQueue) IsEmpty() bool {
	return q.Size() == 0
}
