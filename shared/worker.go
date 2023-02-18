package shared

import (
	"errors"
	"log"
	"sync"
)

type emptyStruct struct{}

// Worker is a routing that runs supervised
type Worker interface {
	Run() error
	Cancel() error
}

// WorkGroup governs a slice of Workers
type WorkGroup struct {
	workers map[Worker]emptyStruct
	wg      sync.WaitGroup
	Done    chan error
}

// Run invokes all Workers and waits for them to finish
func (s *WorkGroup) Run() error {
	var err error
	for k := range s.workers {
		s.wg.Add(1)
		go func(k Worker) {
			defer s.wg.Done()
			err = s.runWorker(k)
		}(k)
	}
	s.wg.Wait()
	return err
}

func (s *WorkGroup) trySendError(err error) {
	select {
	case s.Done <- err:
	default:
		// non blocking send
	}
}

func (s *WorkGroup) runWorker(k Worker) error {
	err := k.Run()
	if err != nil {
		s.trySendError(err)
		log.Println(err)

		err = k.Cancel()
		if err != nil {
			log.Println(err)
		}

		s.Remove(k)
		if len(s.workers) == 0 {
			s.trySendError(errors.New("no more workers running"))
		}
	}

	return err
}

// Cancel interrupts the execution of all workers
func (s *WorkGroup) Cancel() error {
	var err error
	defer close(s.Done)
	for k := range s.workers {
		err = k.Cancel()
		s.Remove(k)
		if err != nil {
			s.trySendError(err)
			log.Println(err)
		}
	}
	return err
}

// Reload calls Reload() on all workers that support it
func (s *WorkGroup) Reload() error {
	var err error
	for k := range s.workers {
		if w, ok := k.(Reloader); ok {
			if e := w.Reload(); e != nil {
				err = e
			}
		}
	}
	return err
}

// NewWorkGroup creates a new empty group of workers
func NewWorkGroup() *WorkGroup {
	s := make(map[Worker]emptyStruct)
	d := make(chan error)
	return &WorkGroup{
		workers: s,
		Done:    d,
	}
}

// Append adds a worker to the group
func (s *WorkGroup) Append(r Worker) {
	if s.workers != nil {
		s.workers[r] = emptyStruct{}
	}
}

// Remove removes a worker from the group
func (s *WorkGroup) Remove(r Worker) {
	delete(s.workers, r)
}
