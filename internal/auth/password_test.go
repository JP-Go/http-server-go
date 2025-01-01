package auth_test

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/JP-Go/http-server-go/internal/auth"
)

func Test_HashPasswordTooLong(t *testing.T) {
	password := "A very long password that has more than 72 characters. It is longer than the 72 characters that bcrypt uses."
	hash := auth.HashPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil || errors.Is(err, bcrypt.ErrPasswordTooLong) {
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

func Test_VerifyPassword(t *testing.T) {
	password := "pass"
	wrongPassword := "other password"
	hash := auth.HashPassword(password)
	err := auth.VerifyPassword(wrongPassword, hash)
	if err == nil || !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Errorf("VerifyPassword(%q, %q) should error", password, hash)
	}
	err = auth.VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("VerifyPassword(%q, %q) should not error: %s", password, hash, err)
	}
}
