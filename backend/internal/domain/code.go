package domain

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// NewTrailCode generates a random trail code in the format TR-XXXX
// where X is a character from the alphabet (no ambiguous O/0/1/I).
func NewTrailCode() string {
	const prefix = "TR-"
	const codeLength = 4

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
