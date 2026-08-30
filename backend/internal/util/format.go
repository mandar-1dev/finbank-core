package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// MaskAccountNumber turns "4821000000004821" into "****4821".
func MaskAccountNumber(number string) string {
	if len(number) <= 4 {
		return "****" + number
	}
	return "****" + number[len(number)-4:]
}

// MaskCardNumber turns "4582000000009214" into "4582 **** **** 9214".
func MaskCardNumber(number string) string {
	if len(number) != 16 {
		return number
	}
	return fmt.Sprintf("%s **** **** %s", number[0:4], number[12:16])
}

// GenerateReference produces a short, unique-enough transaction reference,
// e.g. "TXN-9F3K2A7Q1B". Not cryptographically significant — this is a
// simulation identifier, not a security token.
func GenerateReference(prefix string) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 10)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[n.Int64()]
	}
	return strings.ToUpper(prefix) + "-" + string(b)
}

// GenerateAccountNumber produces a fictional 16-digit account number.
func GenerateAccountNumber() string {
	digits := make([]byte, 16)
	for i := range digits {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		digits[i] = byte('0') + byte(n.Int64())
	}
	return string(digits)
}
