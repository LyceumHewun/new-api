package common

import (
	"strings"

	"github.com/google/uuid"
)

const (
	EmailVerificationPurpose = "v"
	PasswordResetPurpose     = "r"
)

var VerificationValidMinutes = 10

func GenerateVerificationCode(length int) string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	if length == 0 {
		return code
	}
	return code[:length]
}
