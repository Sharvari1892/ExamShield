package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Sharvari1892/examshield/internal/domain"
)



func ComputeHash(prevHash string, payload string, timestamp string) string {
	input := prevHash + payload + timestamp
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func VerifyChain(events []domain.AuditEvent) error {

	for i := 0; i < len(events); i++ {

		expected := ComputeHash(
			events[i].PrevHash,
			events[i].Payload,
			events[i].Timestamp,
		)

		if expected != events[i].CurrentHash {
			return fmt.Errorf("hash mismatch at index %d", i)
		}

		if i > 0 {
			if events[i].PrevHash != events[i-1].CurrentHash {
				return fmt.Errorf("chain linkage broken at index %d", i)
			}
		}
	}

	return nil
}