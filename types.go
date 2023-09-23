package main

import "time"

type Finder interface {
	Get(username string) *Status                                  // If username is empty, return the first available name
	GetByFilter(filter func(index int, name string) bool) *Status // If filter is nil, return the first available name
}

type Status struct {
	Username  string
	Available bool
	drop      [2]time.Time // [0] = start time, [1] = end time. End time is used if the drop time is a range, like NameMC.
}

func (s *Status) First() time.Time {
	return s.drop[0]
}

func (s *Status) Second() time.Time {
	return s.drop[1]
}
