package databases

import "fmt"

type SequenceTaskID struct {
	ID int
}

func (s *SequenceTaskID) String() string {
	return fmt.Sprint(s.ID)
}

type StringTaskID struct {
	ID string
}

func (s *StringTaskID) String() string {
	return s.ID
}
