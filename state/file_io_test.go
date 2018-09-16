package state

import (
	"testing"
)

func TestFileRepositoryImplementsInterface(t *testing.T) {
	var repo Repository = &FileRepository{}
	_ = repo
}
