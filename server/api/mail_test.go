package api

import "testing"

func TestValidEmail(t *testing.T) {
	ok := []string{"a@b.co", "User.Name+tag@example.com"}
	bad := []string{"", "a@b", "a@@b.co", "a @b.co", "no-at.example", "@b.co"}
	for _, email := range ok {
		if !validEmail(email) {
			t.Fatalf("want valid %q", email)
		}
	}
	for _, email := range bad {
		if validEmail(email) {
			t.Fatalf("want invalid %q", email)
		}
	}
	if normalizeEmail(" A@B.CO ") != "a@b.co" {
		t.Fatal("normalize")
	}
}

func TestEnvelopeFrom(t *testing.T) {
	if envelopeFrom("开物录 <noreply@kaiwulu.com>") != "noreply@kaiwulu.com" {
		t.Fatal("named")
	}
	if envelopeFrom("noreply@kaiwulu.com") != "noreply@kaiwulu.com" {
		t.Fatal("plain")
	}
}

func TestResetTokenHash(t *testing.T) {
	token, err := newResetToken()
	if err != nil || len(token) != 64 {
		t.Fatalf("token %q %v", token, err)
	}
	if hashResetToken(token) == token || len(hashResetToken(token)) != 64 {
		t.Fatal("hash")
	}
}
