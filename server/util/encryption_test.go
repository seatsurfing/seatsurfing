package util

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/config"
)

const testCryptKey = "12345678901234567890123456789012" // 32 bytes

func setupCryptKey() {
	GetConfig().CryptKey = testCryptKey
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	setupCryptKey()
	original := "mysecret"
	encrypted, err := EncryptString(original)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}
	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}
	if decrypted != original {
		t.Fatalf("expected %q, got %q", original, decrypted)
	}
}

func TestDecryptStringTooShort(t *testing.T) {
	setupCryptKey()
	// base64 of a 9-byte value — shorter than the 12-byte nonce size
	short := "c2hvcnR2YWw=" // "shortval" (8 bytes)
	_, err := DecryptString(short)
	if err == nil {
		t.Fatal("expected error for too-short ciphertext, got nil")
	}
}

func TestDecryptStringPlaintext(t *testing.T) {
	setupCryptKey()
	// Plain-text string (not base64-encrypted) must not panic and must return an error
	_, err := DecryptString("plaintext-not-encrypted")
	if err == nil {
		t.Fatal("expected error for plain-text input, got nil")
	}
}
