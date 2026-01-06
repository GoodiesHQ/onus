package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

func GenerateCSRFToken() (string, error) {
	var buf [32]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(buf[:]), nil
}

func GenerateNonce() (string, error) {
	var buf [16]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	return base64.URLEncoding.EncodeToString(buf[:]), nil
}

func GenerateULID() (string, error) {
	ms := ulid.Timestamp(time.Now())
	uid, err := ulid.New(ms, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ULID: %w", err)
	}
	return uid.String(), nil
}

func GenerateErrorID() (uuid.UUID, error) {
	return uuid.NewRandom()
}
