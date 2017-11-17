package main

import (
	"strings"
	"sync"
)

type stack struct {
	data []string
	size int
	sync.Mutex
}

func newStack() *stack {
	stk := new(stack)
	return stk
}

func (s *stack) push(t string, tt string) {
	if !s.hasData(t, tt) {
		s.Lock()
		s.data = append(s.data, t+"/"+tt)
		s.size++
		s.Unlock()
	}
}

func (s *stack) hasData(t string, tt string) bool {
	for _, v := range s.data {
		if v == t+"/"+tt {
			return true
		}
	}
	return false
}

func (s *stack) popFor(t string, tt string) {
	s.Lock()
	for i, v := range s.data {
		if v == t+"/"+tt {
			s.data = append(s.data[:i], s.data[i+1:]...)
			s.size--
		}
	}
	s.Unlock()

}

func (s *stack) isEmpty() bool {
	if s.size == 0 {
		return true
	}
	return false
}

func (s *stack) pop() (string, string) {
	var x string
	s.Lock()
	if len(s.data) > 0 {
		x = s.data[len(s.data)-1]
		s.data = s.data[:len(s.data)-1]
		s.size--
	}
	s.Unlock()

	t := strings.Split(x, "/")
	qname, qtype := t[0], t[1]
	return qname, qtype
}
