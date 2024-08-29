// File: utils/utils.go
package utils

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWT claims struct
type JwtClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uuid.UUID, secretKey string, expirationTime time.Duration) (string, error) {
	// Create the JWT claims, which includes the user ID and expiry time
	claims := &JwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response
	return token.SignedString([]byte(secretKey))
}

// Helper function to validate email format
func IsValidEmail(email string) bool {
	// Regular expression for validating an Email
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// Log function to log messages
func Log(v ...interface{}) {
	// Create or open a log file
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new logger
	logger := log.New(file, "", log.LstdFlags)

	// Log the message to the file
	logger.Println(v...)

	// Also print to console
	fmt.Println(v...)
}
