package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	testUUID := uuid.MustParse("39c0998a-a2b4-4c64-8748-210a5c85b6fc")
	pass := "irrigation"
	badPass := "bad"
	expiresIn := time.Second

	tokenString, err := MakeJWT(testUUID, pass, expiresIn)
	if err != nil {
		t.Fatalf("Error generating JWT: %s", err)
	}

	resultUUID, err := ValidateJWT(tokenString, pass)
	if err != nil {
		t.Errorf("Error validating JWT: %s", err)
	} else if resultUUID != testUUID {
		t.Errorf("Validated UUID %v does not equal test UUID %v", resultUUID, testUUID)
	}

	resultUUID, err = ValidateJWT(tokenString, badPass)
	if err == nil {
		t.Errorf("Expected UUID with bad password to fail validation")
	}

	time.Sleep(expiresIn + 100*time.Millisecond)
	resultUUID, err = ValidateJWT(tokenString, pass)
	if err == nil {
		t.Fatalf("UUID %s succesfully validated despite expiring", resultUUID)
	} else if !strings.Contains(err.Error(), "token is expired") {
		t.Fatalf("Expected error to mention token expiry")
	}
}
