package server

import (
	"log"
	"sync"
	"time"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

// WorkerPool manages a pool of workers to process log entries.
type WorkerPool struct {
	logCh       <-chan models.Log
	metrics     *Metrics
	workerCount int
	wg          sync.WaitGroup
}

// NewWorkerPool creates a new WorkerPool.
func NewWorkerPool(logCh <-chan models.Log, metrics *Metrics, workerCount int) *WorkerPool {
	return &WorkerPool{
		logCh:       logCh,
		metrics:     metrics,
		workerCount: workerCount,
	}
}

// Start launches the workers in the pool.
func (p *WorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Stop waits for all workers to finish.
func (p *WorkerPool) Stop() {
	p.wg.Wait()
}

func (p *WorkerPool) worker(id int) {
	log.Printf("worker %d started\n", id)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("worker %d recovered from panic: %v\n", id, r)
		}
		log.Printf("worker %d stopped\n", id)
		p.wg.Done()
	}()

	for logEntry := range p.logCh {
		// Simulate occasional panic
		if logEntry.Message == "panic" {
			panic("simulated worker panic")
		}

		// Simulate processing time
		time.Sleep(2 * time.Second)

		p.metrics.Record(logEntry.Level, logEntry.Service)
		log.Printf("worker %d processed: %+v\n", id, logEntry)
	}
}
