package motadata

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmptyKey      = errors.New("motadata: attribute key must not be empty")
	ErrInvalidKey    = errors.New("motadata: attribute key must contain only alphanumeric characters and dots")
	ErrInvalidFloat  = errors.New("motadata: float value must be finite (not NaN or Inf)")
	ErrEmptySpanName = errors.New("motadata: spanName must not be empty")
)

var validKeyRe = regexp.MustCompile(`^[a-z0-9.]+$`)

const keyPrefix = "apm."

func validateKey(key string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		return "", ErrEmptyKey
	}
	if !validKeyRe.MatchString(k) {
		return "", ErrInvalidKey
	}
	k = strings.TrimPrefix(k, keyPrefix)
	return keyPrefix + k, nil
}
