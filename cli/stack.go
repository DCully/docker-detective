package main

type DirectoryIdStack []int64

func (s *DirectoryIdStack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *DirectoryIdStack) Push(i int64) {
	*s = append(*s, i)
}

func (s *DirectoryIdStack) Pop() (int64, bool) {
	isEmpty := s.IsEmpty()
	if isEmpty {
		return -2, isEmpty
	}
	i := len(*s) - 1
	top := (*s)[i]
	*s = (*s)[:i]
	return top, isEmpty
}
