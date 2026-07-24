package domain

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// NewTrailCode generates a random trail code in the format TR-XXXXXX
// where X is a character from the alphabet (no ambiguous O/0/1/I).
//
// Six characters over a 32-symbol alphabet give 32^6 ≈ 1.07 billion codes, so
// collisions are astronomically unlikely (~0.05% across 1000 codes by the
// birthday bound). Uniqueness is still enforced at persistence via the codigo
// unique constraint; this just keeps random collisions negligible.
func NewTrailCode() string {
	const prefix = "TR-"
	const codeLength = 6

	code := make([]byte, codeLength)
	for i := 0; i < codeLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			// In a production system, this should be handled more gracefully.
			// For now, retry logic is not needed for this implementation.
			panic(err)
		}
		code[i] = alphabet[randomIndex.Int64()]
	}

	return prefix + string(code)
}
