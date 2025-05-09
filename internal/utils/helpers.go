package utils

import "crypto/rand"

var digits = []byte("1234567890")

func GenerateRandomOTP() string {
	const otpLength = 6
	otp := make([]byte, otpLength)

	for i := range otp {
		randomByte := make([]byte, 1)

		for {
			_, err := rand.Read(randomByte)
			if err == nil {
				break
			}
		}

		otp[i] = digits[randomByte[0]%byte(len(digits))]
	}

	return string(otp)
}
