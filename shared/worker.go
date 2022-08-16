package darvaza

import (
	"fmt"
	"log"
	"sync"
)

// Worker is a routing that runs supervised
type Worker interface {
	Run() error
	Cancel() error
}

// WorkGroup governs a slice of Workers
type WorkGroup struct {
	workers map[Worker]struct{}
	wg      sync.WaitGroup
	Done    chan error
}

func (s *WorkGroup) Run() error {
	var err error
	for k := range s.workers {
		s.wg.Add(1)
		go func(k Worker) {
			defer s.wg.Done()
			err = k.Run()
			if err != nil {
				select {
				case s.Done <- err:
				default:
					//non blocking send
				}
				log.Println(err)
				err = k.Cancel()
				if err != nil {
					log.Println(err)
				}
				s.Remove(k)
				if len(s.workers) == 0 {
					select {
					case s.Done <- fmt.Errorf("no more workers running"):
					default:
						//non blocking send
					}
				}
			}
		}(k)
	}
	s.wg.Wait()
	return err
}

func (s *WorkGroup) Cancel() error {
	var err error
	defer close(s.Done)
	for k := range s.workers {
		err = k.Cancel()
		s.Remove(k)
		if err != nil {
			select {
			case s.Done <- err:
				log.Println(err)
			default:
				//non blocking send
			}
		}
	}
	return err
}

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

func NewWorkGroup() *WorkGroup {
	s := make(map[Worker]struct{})
	d := make(chan error)
	return &WorkGroup{
		workers: s,
		Done:    d,
	}
}

func (s *WorkGroup) Append(r Worker) {
	if s.workers != nil {
		s.workers[r] = struct{}{}
	}
}

func (s *WorkGroup) Remove(r Worker) {
	delete(s.workers, r)
}
