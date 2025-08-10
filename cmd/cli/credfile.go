package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/awryme/reddit-exporter/pkg/jsonfile"
	"golang.org/x/term"
)

type Creds struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func readCredentials(clientID string) (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	if clientID == "" {
		fmt.Print("Enter client_id: ")
		id, err := reader.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		clientID = id
	}

	fmt.Print("Enter client_secret: ")
	clientSecred, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}

	password := string(clientSecred)
	return strings.TrimSpace(clientID), strings.TrimSpace(password), nil
}

func SaveCredsToFile(file string, clientID string) (Creds, error) {
	id, secret, err := readCredentials(clientID)
	if err != nil {
		return Creds{}, fmt.Errorf("read creds from stdin: %w", err)
	}

	creds := Creds{
		ClientID:     id,
		ClientSecret: secret,
	}
	err = jsonfile.Write(file, creds)
	return creds, err
}

func ReadCredsFromFile(file string) (Creds, error) {
	return jsonfile.Read[Creds](file)
}
