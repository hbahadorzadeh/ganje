package main

import (
    "fmt"
    "os"
    "strings"
    "time"

    jwt "github.com/golang-jwt/jwt/v5"
)

// Minimal HS256 JWT generator compatible with internal/auth.Claims
// Usage: go run scripts/it/jwt.go <secret> <username> <email> <realm1,realm2,...>
func main() {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "usage: %s <secret> <username> <email> <realmsCSV>\n", os.Args[0])
		os.Exit(2)
	}
	secret := os.Args[1]
	username := os.Args[2]
	email := os.Args[3]
	realms := strings.Split(os.Args[4], ",")

	claims := jwt.MapClaims{
		"username": username,
		"email":    email,
		"realms":   realms,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Output as `Bearer <token>` to ease curl usage when command-substituted
	fmt.Println("Bearer " + signed)
}
