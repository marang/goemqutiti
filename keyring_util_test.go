//go:build manual

package main

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

// ExampleGet demonstrates retrieving a password from the real keyring.
func ExampleSet_manual() {
	// Define test data
	service := "ExampleService"
	username := "exampleuser"
	password := "examplepassword"

	// Pre-store a password in the real keyring
	err := keyring.Set(service, username, password)
	if err != nil {
		fmt.Println("Setup Error:", err)
		os.Exit(1)
	}

	fmt.Println("Set Password for user", username)

	// Retrieve the password from the real keyring
	retrievedPassword, err := keyring.Get(service, username)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Retrieved Password:", retrievedPassword)

	// Clean up: Remove the password from the keyring after the test

	if err = keyring.Delete(service, username); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Retrieve the password from the real keyring
	if retrievedPassword, err = keyring.Get(service, username); retrievedPassword != "" || err == nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Output:
	// Set Password for user exampleuser
	// Retrieved Password: examplepassword
}
