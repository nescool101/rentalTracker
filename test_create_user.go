package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	// The password we want to use
	plainPassword := "password123"

	// Base64 encode it
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(plainPassword))

	// Print both for reference
	fmt.Printf("Plain password: %s\n", plainPassword)
	fmt.Printf("Base64 encoded: %s\n", encodedPassword)

	// Also create a curl command to test login
	curlCmd := fmt.Sprintf("curl 'http://localhost:8080/api/users/login' -X POST -H 'Content-Type: application/json' --data-raw '{\"email\":\"test@example.com\",\"password\":\"%s\"}'", encodedPassword)

	fmt.Printf("\nTest login with:\n%s\n", curlCmd)

	// Save the curl command to a file for easy execution
	os.WriteFile("test_login.sh", []byte(curlCmd), 0755)
	fmt.Println("\nSaved curl command to test_login.sh")
}
