package passwd

import "golang.org/x/crypto/bcrypt"

func HashPasswd(pwd string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pwd), 8)
}

func CheckPasswd(pwd string, pwHashed []byte) error {
	return bcrypt.CompareHashAndPassword(pwHashed, []byte(pwd))
}
