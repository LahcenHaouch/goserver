package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	if _, err := MakeJWT(uuid.New(), "Lahcen", time.Hour); err != nil {
		t.Fatalf("error creating jwt token: %q", err)
	}
}

func TestValidateJWT(t *testing.T) {
	id := uuid.New()
	secret := "Lahcen"
	token, err := MakeJWT(id, secret, time.Hour)
	if err != nil {
		t.Fatalf("error creating jwt token: %q", err)
	}

	if _, err := ValidateJWT(token, secret); err != nil {
		t.Fatalf("token not valid")
	}
}

// [todo] add test case for expired tokens
