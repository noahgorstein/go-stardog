// The purpose of this example is to demonstrate how to work with database metadata (a.k.a configuration options).
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("Endpoint (leave empty for http://localhost:5820): ")
	endpoint, _ := r.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = "http://localhost:5820"
	}

	fmt.Print("Username (leave empty for admin): ")
	username, _ := r.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	fmt.Print("Password (leave empty for admin): ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	if password == "" {
		password = "admin"
	}
	fmt.Println()

	basicAuthTransport := stardog.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client, err := stardog.NewClient(endpoint, basicAuthTransport.Client())
	if err != nil {
		log.Fatalf("unable to create stardog client: %v", err)
	}

	fmt.Print("Database to set 'search.enabled=true' for: ")
	database, _ := r.ReadString('\n')
	database = strings.TrimSpace(database)

	fmt.Println("Offlining the database...")
	_, err = client.DatabaseAdmin.Offline(context.Background(), database)
	if err != nil {
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}
	fmt.Println("Database offlined successfully.")

	fmt.Println("Setting the database configuration option 'spatial.enabled=true'")
	setOptions := map[string]interface{}{
		"search.enabled": true,
	}
	_, err = client.DatabaseAdmin.SetMetadata(context.Background(), database, setOptions)
	if err != nil {
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}

	configOptions := []string{"search.enabled"}
	opts, _, err := client.DatabaseAdmin.Metadata(context.Background(), database, configOptions)
	if err != nil {
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}

	fmt.Println("-----DATABASE OPTIONS----")
	for key, value := range opts {
		fmt.Printf("OPTION: %s | VALUE: %v\n", key, value)
	}
	fmt.Println("----------------")

	fmt.Printf("Onlining the database %s...\n", database)
	_, err = client.DatabaseAdmin.Online(context.Background(), database)
	if err != nil {
		fmt.Printf("Unable to online database \"%s\"\n", database)
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}
	fmt.Println("Database onlined successfully.")
}
