package wikibfs

import "github.com/google/uuid"

type algorithm struct {
}

func newAlgorithm() *algorithm {
	return &algorithm{}
}

func (a *algorithm) findShortestPath(taskID uuid.UUID, from, to string) ([]string, error) {
	return []string{from, to}, nil
}
