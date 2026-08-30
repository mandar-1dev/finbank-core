package util

import "testing"

func TestMaskAccountNumber(t *testing.T) {
	got := MaskAccountNumber("4821000000004821")
	want := "****4821"
	if got != want {
		t.Errorf("MaskAccountNumber() = %s, want %s", got, want)
	}
}

func TestMaskCardNumber(t *testing.T) {
	got := MaskCardNumber("4582000000009214")
	want := "4582 **** **** 9214"
	if got != want {
		t.Errorf("MaskCardNumber() = %s, want %s", got, want)
	}
}

func TestGenerateReferenceIsUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		ref := GenerateReference("TXN")
		if seen[ref] {
			t.Fatalf("duplicate reference generated: %s", ref)
		}
		seen[ref] = true
	}
}

func TestGenerateAccountNumberLength(t *testing.T) {
	num := GenerateAccountNumber()
	if len(num) != 16 {
		t.Errorf("expected 16-digit account number, got %d digits: %s", len(num), num)
	}
}
