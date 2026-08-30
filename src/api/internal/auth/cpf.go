package auth

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidCPF = errors.New("invalid cpf")

// NormalizeCPF strips non-digit characters and validates the CPF checksum,
// returning the digits-only representation.
func NormalizeCPF(cpf string) (string, error) {
	var digits strings.Builder
	for _, r := range cpf {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}

	normalized := digits.String()

	if len(normalized) != 11 || allDigitsEqual(normalized) {
		return "", ErrInvalidCPF
	}

	if !validCPFChecksum(normalized) {
		return "", ErrInvalidCPF
	}

	return normalized, nil
}

func allDigitsEqual(cpf string) bool {
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			return false
		}
	}
	return true
}

func validCPFChecksum(cpf string) bool {
	digits := make([]int, len(cpf))
	for i, r := range cpf {
		d, err := strconv.Atoi(string(r))
		if err != nil {
			return false
		}
		digits[i] = d
	}

	firstCheck := cpfCheckDigit(digits[:9], 10)
	if firstCheck != digits[9] {
		return false
	}

	secondCheck := cpfCheckDigit(digits[:10], 11)
	return secondCheck == digits[10]
}

func cpfCheckDigit(digits []int, firstWeight int) int {
	sum := 0
	weight := firstWeight
	for _, d := range digits {
		sum += d * weight
		weight--
	}

	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
