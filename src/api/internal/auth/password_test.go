package auth_test

import (
	"testing"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
)

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"too short with special char", "Ab1!", true},
		{"long but no special char", "abcdefghij", true},
		{"valid", "abcdefgh1!", false},
		{"exactly 9 chars valid", "abcdefg1!", false},
		{"9 chars no special", "abcdefghi", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := auth.ValidatePassword(tc.password)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for password %q, got nil", tc.password)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for password %q, got %v", tc.password, err)
			}
		})
	}
}

func TestHashAndCompatiblePassword(t *testing.T) {
	hash, err := auth.HashPassword("abcdefg1!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !auth.ComparePassword(hash, "abcdefg1!") {
		t.Fatal("expected password to match hash")
	}

	if auth.ComparePassword(hash, "wrongpassword1!") {
		t.Fatal("expected wrong password not to match hash")
	}
}
