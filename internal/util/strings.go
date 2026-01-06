package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ExtractDomain(email string) (string, error) {
	at := strings.LastIndex(email, "@")
	if at == -1 || at == len(email)-1 {
		return "", fmt.Errorf("invalid email address")
	}

	domain := strings.TrimSpace(strings.ToLower(email[at+1:]))
	if domain == "" {
		return "", fmt.Errorf("invalid email address")
	}

	return domain, nil
}

func FirstNonEmpty(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}

func IsValidEmail(email string) bool {
	if strings.Count(email, "@") != 1 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at == -1 || at == 0 || at == len(email)-1 {
		return false
	}
	return true
}

func ParseBool(str string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(str)) {
	case "1", "true", "t", "yes", "y", "on":
		return true, nil
	case "0", "false", "f", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean string: %q", str)
	}
}

func ParseInt[T ~int](str string, def T) T {
	str = strings.TrimSpace(str)
	if str == "" {
		return def
	}

	i, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return def
	}
	return T(i)
}
