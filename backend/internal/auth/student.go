package auth

import (
	"crypto/rand"
	"math/big"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// alphabet excludes ambiguous chars (i, l, 1, o, 0) for readability.
const credAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// UsernameSlug normalizes a name into the stable part of a login: lowercase,
// accents removed, runs of non-alphanumerics collapsed to a single dot.
func UsernameSlug(nome string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, err := transform.String(t, nome)
	if err != nil {
		normalized = nome
	}
	normalized = strings.ToLower(normalized)
	var b strings.Builder
	lastDot := true // avoid leading dot
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDot = false
		} else if !lastDot {
			b.WriteByte('.')
			lastDot = true
		}
	}
	return strings.Trim(b.String(), ".")
}

func randomFrom(alphabet string, n int) (string, error) {
	b := make([]byte, n)
	max := big.NewInt(int64(len(alphabet)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[idx.Int64()]
	}
	return string(b), nil
}

// RandomSuffix returns a 4-char unambiguous suffix for username uniqueness.
func RandomSuffix() (string, error) { return randomFrom(credAlphabet, 4) }

// GenerateInitialPassword returns an 8-char unambiguous initial password.
func GenerateInitialPassword() (string, error) { return randomFrom(credAlphabet, 8) }
