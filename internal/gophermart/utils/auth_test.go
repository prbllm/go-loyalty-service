package utils

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "secret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if !CheckPassword(hash, password) {
		t.Fatalf("password should be valid")
	}

	if CheckPassword(hash, "wrong") {
		t.Fatalf("expected invalid password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	userID := int64(42)

	token, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	parsedID, err := ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}

	if parsedID != userID {
		t.Fatalf("expected userID %d, got %d", userID, parsedID)
	}
}
