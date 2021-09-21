package types

import (
	"fmt"
	"log"
	"sync"
)

// Runner is an interface which is implemented by all proxies
type Runner interface {
	Run() error
	Cancel() error
	Reload() error
}

// Server governs a slice of Runners
type Server struct {
	Servers map[Runner]struct{}
	wg      sync.WaitGroup
	Done    chan error
}

func (s *Server) Run() error {
	var err error
	for k, _ := range s.Servers {
		s.wg.Add(1)
		go func(k Runner) {
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
				if len(s.Servers) == 0 {
					select {
					case s.Done <- fmt.Errorf("no more servers running"):
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

func (s *Server) Cancel() error {
	var err error
	defer close(s.Done)
	for k, _ := range s.Servers {
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

func (s *Server) Reload() error {
	return nil
}

func NewServer() *Server {
	s := make(map[Runner]struct{})
	d := make(chan error)
	return &Server{
		Servers: s,
		Done:    d,
	}
}

func (s *Server) Append(r Runner) {
	if s.Servers != nil {
		s.Servers[r] = struct{}{}
	}
}

func (s *Server) Remove(r Runner) {

	delete(s.Servers, r)

}
