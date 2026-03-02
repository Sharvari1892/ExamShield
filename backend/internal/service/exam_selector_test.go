package service

import (
	"testing"

	"github.com/Sharvari1892/examshield/internal/repository"
)

func TestDeterministicSelection(t *testing.T) {
	questions := []repository.Question{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	}

	a := DeterministicSelect("student1", "exam1", questions)
	b := DeterministicSelect("student1", "exam1", questions)

	for i := range a {
		if a[i].ID != b[i].ID {
			t.Fatalf("selection not deterministic")
		}
	}
}

func TestDifferentUsersDifferentSelection(t *testing.T) {
	questions := []repository.Question{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	}

	a := DeterministicSelect("student1", "exam1", questions)
	b := DeterministicSelect("student2", "exam1", questions)

	same := true
	for i := range a {
		if a[i].ID != b[i].ID {
			same = false
			break
		}
	}

	if same {
		t.Fatalf("different users should not get same selection")
	}
}
