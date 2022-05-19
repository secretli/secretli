package internal

import (
	"fmt"
	"github.com/mattn/go-tty"
	"io"
	"log"
	"os"
	"strings"
)

func ReadFromStdin() (string, error) {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func GetPasswordFromTerminalOrDie() string {
	tty, err := tty.Open()
	if err != nil {
		log.Fatalln(err)
	}
	defer tty.Close()

	fmt.Fprint(tty.Output(), "Enter Password: ")
	password, err := tty.ReadPassword()
	if err != nil {
		log.Fatalln(err)
	}

	return strings.TrimSpace(password)
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
