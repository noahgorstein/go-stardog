// The purpose of this example is to demonstrate how to import namespaces
// contained in an RDF file into a database.
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

	fmt.Print("Database to import schema.org namespace into: ")
	database, _ := r.ReadString('\n')
	database = strings.TrimSpace(database)

	rdfFile, err := os.Open("namespaces.ttl")
	if err != nil {
		log.Fatalf("unable to open data file to be imported: %v", err)
	}

	importNamespacesResponse, _, err := client.DatabaseAdmin.ImportNamespaces(context.Background(), database, rdfFile)
	if err != nil {
		fmt.Println("unable to import namespaces")
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}

	fmt.Println()
	fmt.Printf("Successfully imported namespace into database: \"%s\"\n", database)
	fmt.Printf("Number of namespaces imported: %d\n", importNamespacesResponse.NumberImportedNamespaces)
	fmt.Println("-------Updated Namespaces------")
	for _, ns := range importNamespacesResponse.UpdatedNamespaces {
		fmt.Println(ns)
	}
}
