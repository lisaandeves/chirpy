package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(pwd string) (string, error) {
	hashed, err := argon2id.CreateHash(pwd, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashed, nil
}

func CheckPasswordHash(pwd, hash string) (bool, error) {
	ok, err := argon2id.ComparePasswordAndHash(pwd, hash)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	issuer := "chirpy-access"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	callback := func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("Incorrect signing method %s", t.Method.Alg())
		}
		return []byte(tokenSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, callback)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error validating user: %s", err)
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error validating user: %s", err)
	} else if issuer != "chirpy-access" {
		return uuid.Nil, errors.New("Invalid issuer")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error validating user: %s", err)
	}
	id, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error validating user: %s", err)
	}

	return id, nil
}
