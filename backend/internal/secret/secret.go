// Package secret provides authenticated encryption (AES-256-GCM) for small
// secrets like API tokens, so they can be stored encrypted at rest.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// Box holds an AES-256-GCM AEAD built from a 32-byte key.
type Box struct {
	aead cipher.AEAD
}

// NewBox decodes a base64 (StdEncoding) 32-byte key and builds a Box.
func NewBox(base64Key string) (*Box, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("secret: key must be 32 bytes (AES-256)")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{aead: aead}, nil
}

// Seal encrypts plaintext, returning the ciphertext and the random nonce used.
func (b *Box) Seal(plaintext []byte) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = b.aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Open decrypts ciphertext using its nonce. Returns an error if the key is
// wrong or the data was tampered with.
func (b *Box) Open(ciphertext, nonce []byte) ([]byte, error) {
	return b.aead.Open(nil, nonce, ciphertext, nil)
}
