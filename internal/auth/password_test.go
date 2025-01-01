package auth_test

import (
	"golang.org/x/crypto/bcrypt"
	"testing"

	"github.com/JP-Go/http-server-go/internal/auth"
)

func Test_HashEmptyPassword(t *testing.T) {
	password := ""
	hash := auth.HashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("HashPassword(%q) hash is empty: error %s", password, err)
	}
}

func Test_HashPasswordTooLong(t *testing.T) {
	password := "A very long password that has more than 72 characters. It is longer than the 72 characters that bcrypt uses."
	hash := auth.HashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		t.Errorf("HashPassword(%q) should error", password)
	}
}

func Test_HashPassword(t *testing.T) {
	password := "pass"
	hash := auth.HashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("HashPassword(%q) hash pass errors: error %s", password, err)
	}
}

func Test_VerifyPasswordRightPassword(t *testing.T) {
	password := "pass"
	hash := auth.HashPassword(password)
	err := auth.VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("VerifyPassword(%q, %q) errors: error %s", password, hash, err)
	}
}

func Test_VerifyPasswordWrongPassword(t *testing.T) {
	password := "pass"
	hash := auth.HashPassword(password)
	err := auth.VerifyPassword("other password", hash)
	if err == nil {
		t.Errorf("VerifyPassword(%q, %q) should error", password, hash)
	}
}
