package apihelpers

import (
	"net/mail"
	"regexp"
	"unicode"
)

var usernameRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]{2,49}$")

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func IsValidPassword(passwd string) bool {
	if len(passwd) < 12 {
		return false
	}
	if len(passwd) >= 50 {
		return false
	}
	hasUpperCase := false
	hasLowerCase := false
	hasNumeric := false
	hasSpecial := false
	for _, r := range passwd {
		isUpper := unicode.IsUpper(r)
		isLower := unicode.IsLower(r)
		isNumeric := unicode.IsNumber(r)
		hasUpperCase = hasUpperCase || isUpper
		hasLowerCase = hasLowerCase || isLower
		hasNumeric = hasNumeric || isNumeric
		hasSpecial = hasSpecial || !(isUpper || isLower || isNumeric || unicode.IsSpace(r) || unicode.IsControl(r))
	}
	return hasUpperCase && hasLowerCase && hasNumeric && hasSpecial
}
