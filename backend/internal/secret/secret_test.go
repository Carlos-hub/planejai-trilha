package secret

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func testKey(t *testing.T) string {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(k)
}

func TestSealOpenRoundTrip(t *testing.T) {
	b, err := NewBox(testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("sk-ant-super-secret-123")
	ct, nonce, err := b.Seal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(ct, msg) {
		t.Fatal("ciphertext must not equal plaintext")
	}
	got, err := b.Open(ct, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatalf("open = %q, want %q", got, msg)
	}
}

func TestSealUsesFreshNonce(t *testing.T) {
	b, _ := NewBox(testKey(t))
	_, n1, _ := b.Seal([]byte("same"))
	_, n2, _ := b.Seal([]byte("same"))
	if bytes.Equal(n1, n2) {
		t.Fatal("two Seal calls must use different nonces")
	}
}

func TestOpenWrongKeyFails(t *testing.T) {
	b1, _ := NewBox(testKey(t))
	b2, _ := NewBox(testKey(t))
	ct, nonce, _ := b1.Seal([]byte("secret"))
	if _, err := b2.Open(ct, nonce); err == nil {
		t.Fatal("Open with wrong key must fail")
	}
}

func TestNewBoxRejectsBadKey(t *testing.T) {
	if _, err := NewBox("not-base64!!!"); err == nil {
		t.Fatal("invalid base64 must error")
	}
	short := base64.StdEncoding.EncodeToString(make([]byte, 16))
	if _, err := NewBox(short); err == nil {
		t.Fatal("16-byte key must error (need 32)")
	}
}
