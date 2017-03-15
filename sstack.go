package main

import (
	"strings"
	"sync"
)

type Stack struct {
	data []string
	size int
	sync.Mutex
}

func NewStack() *Stack {
	stk := new(Stack)
	return stk
}

func (s *Stack) Push(t string, tt string) {
	if !s.hasData(t, tt) {
		s.Lock()
		s.data = append(s.data, t+"/"+tt)
		s.size++
		s.Unlock()
	}
}

func (s *Stack) hasData(t string, tt string) bool {
	for _, v := range s.data {
		if v == t+"/"+tt {
			return true
		}
	}
	return false
}

func (s *Stack) PopFor(t string, tt string) {
	s.Lock()
	for i, v := range s.data {
		if v == t+"/"+tt {
			s.data = append(s.data[:i], s.data[i+1:]...)
			s.size--
		}
	}
	s.Unlock()

}

func (s *Stack) Size() int {
	return s.size
}

func (s *Stack) IsEmpty() bool {
	if s.size == 0 {
		return true
	}
	return false
}

func (s *Stack) Pop() (string, string) {
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
