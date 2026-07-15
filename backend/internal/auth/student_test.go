package auth

import (
	"regexp"
	"strings"
	"testing"
)

func TestUsernameSlug(t *testing.T) {
	cases := map[string]string{
		"Ana Clara":      "ana.clara",
		"João da Silva":  "joao.da.silva",
		"MARIA":          "maria",
		"  Zé  Ninguém ": "ze.ninguem",
	}
	for in, want := range cases {
		if got := UsernameSlug(in); got != want {
			t.Errorf("UsernameSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRandomSuffixAndPassword(t *testing.T) {
	ok := regexp.MustCompile(`^[a-hj-np-z2-9]+$`)
	suf, err := RandomSuffix()
	if err != nil || len(suf) != 4 || !ok.MatchString(suf) {
		t.Fatalf("suffix=%q err=%v", suf, err)
	}
	pw, err := GenerateInitialPassword()
	if err != nil || len(pw) != 8 || !ok.MatchString(pw) {
		t.Fatalf("pw=%q err=%v", pw, err)
	}
	if strings.ContainsAny(pw, "il1o0") {
		t.Fatalf("ambiguous chars in %q", pw)
	}
}
