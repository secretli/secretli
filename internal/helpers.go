package internal

import (
	"encoding/base64"
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

func B64Encode(input []byte) string {
	return base64.RawURLEncoding.EncodeToString(input)
}

func B64Decode(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}
