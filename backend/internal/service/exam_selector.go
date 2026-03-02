package service

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"

	"github.com/Sharvari1892/examshield/internal/repository"
)

func DeterministicSelect(studentID, examID string, questions []repository.Question) []repository.Question {
	hash := sha256.Sum256([]byte(studentID + examID))
	seed := int64(binary.BigEndian.Uint64(hash[:8]))

	rng := rand.New(rand.NewSource(seed))

	shuffled := make([]repository.Question, len(questions))
	copy(shuffled, questions)

	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}
