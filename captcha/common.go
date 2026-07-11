package captcha

import "math/rand"

type value struct {
	Code    string
	Carrier string
}

func generateCode() string {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}
