package internal

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

func GetPasswordFromTerminalOrDie() string {
	fmt.Print("Enter Password: ")
	password, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()

	if err != nil {
		os.Exit(2)
	}

	return strings.TrimSpace(string(password))
}

func SetupStore(baseUrl string) (*HTTPRemoteStore, error) {
	var baseUrlFunc ClientOptionFunc
	if baseUrl != "" {
		baseUrlFunc = WithBaseURL(baseUrl)
	}

	client, err := NewClient(baseUrlFunc)
	if err != nil {
		return nil, err
	}

	return NewHTTPRemoteStore(client), nil
}
