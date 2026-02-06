package auth

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GenerateSecret creates a new TOTP secret and returns the secret string and provisioning URI.
func GenerateSecret(issuer, account string) (secret string, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// ValidateTOTP validates a TOTP code against the given secret.
func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
