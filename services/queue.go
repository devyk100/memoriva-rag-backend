package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type StudySessionJob struct {
	SessionID string
	Timestamp time.Time
}

type QueueService struct {
	jobs        chan StudySessionJob
	workers     int
	ragService  *RAGService
	dbService   *DatabaseService
	workerGroup sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewQueueService(workers int, ragService *RAGService, dbService *DatabaseService) *QueueService {
	ctx, cancel := context.WithCancel(context.Background())

	return &QueueService{
		jobs:       make(chan StudySessionJob, 100), // Buffer for 100 jobs
		workers:    workers,
		ragService: ragService,
		dbService:  dbService,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (q *QueueService) Start() {
	log.Printf("Starting queue service with %d workers", q.workers)

	for i := 0; i < q.workers; i++ {
		q.workerGroup.Add(1)
		go q.worker(i)
	}
}

func (q *QueueService) Stop() {
	log.Println("Stopping queue service...")
	q.cancel()
	close(q.jobs)
	q.workerGroup.Wait()
	log.Println("Queue service stopped")
}

func (q *QueueService) EnqueueStudySession(sessionID string) error {
	job := StudySessionJob{
		SessionID: sessionID,
		Timestamp: time.Now(),
	}

	select {
	case q.jobs <- job:
		log.Printf("Enqueued study session: %s", sessionID)
		return nil
	case <-q.ctx.Done():
		return q.ctx.Err()
	default:
		log.Printf("Queue is full, rejecting session: %s", sessionID)
		return ErrQueueFull
	}
}

func (q *QueueService) worker(workerID int) {
	defer q.workerGroup.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case job, ok := <-q.jobs:
			if !ok {
				log.Printf("Worker %d: channel closed, exiting", workerID)
				return
			}

			log.Printf("Worker %d processing session: %s", workerID, job.SessionID)
			q.processStudySession(workerID, job)

		case <-q.ctx.Done():
			log.Printf("Worker %d: context cancelled, exiting", workerID)
			return
		}
	}
}

func (q *QueueService) processStudySession(workerID int, job StudySessionJob) {
	sessionID := job.SessionID

	// Process with RAG service (which handles all the logic internally)
	err := q.ragService.ProcessStudySession(sessionID)
	if err != nil {
		log.Printf("Worker %d: RAG processing failed for session %s: %v", workerID, sessionID, err)
		return
	}

	log.Printf("Worker %d: Successfully processed study session %s", workerID, sessionID)
}

// Custom errors
var (
	ErrQueueFull = fmt.Errorf("queue is full")
)
