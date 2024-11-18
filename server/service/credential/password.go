package credential

import (
	"math/rand"
	"time"
)

const (
	passwordLength = 20
	charset        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?/|"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func GeneratePassword() string {
	password := make([]byte, passwordLength)
	for i := range password {
		password[i] = charset[r.Intn(len(charset))]
	}

	return string(password)
}
