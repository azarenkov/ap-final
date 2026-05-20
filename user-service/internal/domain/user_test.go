package domain

import (
	"errors"
	"testing"
)

func TestNormaliseEmail(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr error
	}{
		{"alice@example.com", "alice@example.com", nil},
		{"  ALICE@EXAMPLE.com  ", "alice@example.com", nil},
		{"not-an-email", "", ErrInvalidEmail},
		{"", "", ErrInvalidEmail},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := NormaliseEmail(tc.in)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err: got %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("value: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("short password should be weak, got %v", err)
	}
	if err := ValidatePassword("longenough"); err != nil {
		t.Fatalf("10-char password should pass, got %v", err)
	}
}

func TestValidateName(t *testing.T) {
	if err := ValidateName(""); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("empty should fail, got %v", err)
	}
	if err := ValidateName("   "); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("whitespace-only should fail, got %v", err)
	}
	if err := ValidateName("Jane Doe"); err != nil {
		t.Fatalf("got %v", err)
	}
}
