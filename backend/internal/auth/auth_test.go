package auth

import "testing"

func TestHashAndCheck(t *testing.T) {
	h, err := HashPassword("segredo123")
	if err != nil { t.Fatal(err) }
	if !CheckPassword(h, "segredo123") { t.Fatal("should match") }
	if CheckPassword(h, "errado") { t.Fatal("should not match") }
}

func TestSessionIDUnique(t *testing.T) {
	a, _ := NewSessionID()
	b, _ := NewSessionID()
	if a == b || len(a) < 20 { t.Fatalf("bad session ids %q %q", a, b) }
}
