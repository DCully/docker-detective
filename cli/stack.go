package main

type DirectoryIdStack []int64

func (s *DirectoryIdStack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *DirectoryIdStack) Push(i int64) {
	*s = append(*s, i)
}

func (s *DirectoryIdStack) Pop() (int64, bool) {
	if s.IsEmpty() {
		return -2, false
	}
	i := len(*s) - 1
	top := (*s)[i]
	*s = (*s)[:i]
	return top, true
}
