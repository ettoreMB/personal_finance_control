package auth_test

import (
	"testing"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
)

func TestNormalizeAndValidateCPF(t *testing.T) {
	cases := []struct {
		name    string
		cpf     string
		want    string
		wantErr bool
	}{
		{"valid formatted", "529.982.247-25", "52998224725", false},
		{"valid digits only", "52998224725", "52998224725", false},
		{"invalid checksum", "52998224726", "", true},
		{"all same digits", "11111111111", "", true},
		{"wrong length", "123456789", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := auth.NormalizeCPF(tc.cpf)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for cpf %q, got nil", tc.cpf)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
