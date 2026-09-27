package auth

import "github.com/alexedwards/argon2id"

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
