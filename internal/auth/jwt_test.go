package auth_test

import (
	"testing"
	"time"

	"github.com/JP-Go/http-server-go/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "secret"

func Test_JWTCanMake(t *testing.T) {
	userID := uuid.New()
	expirationTime := time.Second * 5

	signedString, err := auth.MakeJWT(userID, testSecret, expirationTime)
	if err != nil || signedString == "" {
		t.Fatalf("Could not sign token due to %s", err)
	}

	token, err := jwt.Parse(signedString, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	}, jwt.WithIssuer(auth.DefaultIssuer))
	if !token.Valid {
		t.Fatalf("Created an invalid token")
	}
	if token.Method.Alg() != "HS256" {
		t.Fatalf("Invalid alg for singing. Expected Hs256, got: %s", token.Method.Alg())
	}
	issuedAt, err := token.Claims.GetIssuedAt()
	if err != nil {
		t.Fatal("Should not be able to issue a token without issued at")
	}
	expiresIn, err := token.Claims.GetExpirationTime()
	if err != nil {
		t.Fatal("Should not be able to issue a token without expires at")
	}
	if expiresIn.Time.Sub(issuedAt.Time) != expirationTime {
		t.Fatalf("Created a token where the expirationTime was diferent than the provided\n Expires: %v, Issued: %v", expiresIn.Time, issuedAt.Time)
	}

	if expiresIn.Time.Sub(issuedAt.Time) != expirationTime {
		t.Fatal("Created a token where the expirationTime was diferent than the provided")
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		t.Fatal("Should not be able to issue a token with no issuer")
	}
	if issuer != auth.DefaultIssuer {
		t.Fatal("Created a token where the expirationTime was diferent than the provided")
	}

}

func Test_ValidateJWT_TokenExpires(t *testing.T) {
	userID := uuid.New()
	t.Parallel()
	expirationTime := time.Second * 1

	signedString, err := auth.MakeJWT(userID, testSecret, expirationTime)
	timer := time.NewTimer(expirationTime + time.Second)
	<-timer.C
	_, err = auth.ValidateJWT(signedString, testSecret)
	if err == nil {
		t.Fatal("Should fail in expired token")
	}

}

func Test_JWTVerify(t *testing.T) {
	userID := uuid.New()
	expirationTime := time.Second * 5

	signedString, err := auth.MakeJWT(userID, testSecret, expirationTime)
	if err != nil || signedString == "" {
		t.Fatalf("Could not sign token due to %s", err)
	}
	possibleUserID, err := auth.ValidateJWT(signedString, testSecret)
	if err != nil {
		t.Fatalf("Could not validate jwt due to: %s", err)
	}

	if possibleUserID != userID {
		t.Fatalf("Wrong user id returned. Expected %v, got %v", userID, possibleUserID)
	}

}
