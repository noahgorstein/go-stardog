// The purpose of this example is to demonstrate how to export data from a Stardog database.
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
		log.Fatalf("unable to create Stardog client: %v", err)
	}

	fmt.Print("Database to export default graph of: ")
	database, _ := r.ReadString('\n')
	database = strings.TrimSpace(database)

	opts := &stardog.ExportDataOptions{
		NamedGraph: []string{"tag:stardog:api:context:default"},
		Format:     stardog.RDFFormatTurtle,
	}

	buf, _, err := client.DatabaseAdmin.ExportData(context.Background(), database, opts)
	if err != nil {
		fmt.Println("unable to export data")
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}
	fmt.Printf("Successfully exported database: \"%s\"\n", database)
	fmt.Println("-------DATA (Turtle Format)-------")
	if buf != nil {
		fmt.Println(buf.String())
	}
}
