package databases

import "fmt"

type IntTaskID struct {
	ID int
}

func (s *IntTaskID) String() string {
	return fmt.Sprint(s.ID)
}

type StringTaskID struct {
	ID string
}

func (s *StringTaskID) String() string {
	return s.ID
}
