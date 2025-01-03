package auth

import (
	"errors"
	"net/http"
	"strings"
)

func getTokenFromHeader(headers http.Header, prefix string) (string, error) {
	headerValue := headers.Get("Authorization")
	if headerValue == "" {
		return "", errors.New("Authorization header not present")
	}
	if !strings.Contains(headerValue, prefix) {
		return "", errors.New("Invalid token type")
	}
	return strings.TrimPrefix(headerValue, prefix+" "), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	return getTokenFromHeader(headers, "Bearer")
}

func GetAPIKey(headers http.Header) (string, error) {
	return getTokenFromHeader(headers, "ApiKey")
}
