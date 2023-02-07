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

//  fmt.Println("Adding ds...")
//  opts := map[string]interface{}{
//    "jdbc.url": "jdbc:postgresql://localhost:5432/employees",
//    "jdbc.driver": "org.postgresql.Driver",
//  }
//  _, err = client.DataSource.Add(context.Background(), "postgres2", opts)
//	if err != nil {
//		var stardogErr *stardog.ErrorResponse
//		if errors.As(err, &stardogErr) {
//			log.Fatalf("stardog error occurred: %v", err)
//		}
//		log.Fatalf("non-stardog error occurred: %v", err)
//	}

  _, err = client.DataSource.RefreshMetadata(context.Background(), "postgres", nil)
	if err != nil {
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}

  ds, _, err := client.DataSource.List(context.Background())
	if err != nil {
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}
	fmt.Println(ds)
}
