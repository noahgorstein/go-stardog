// The purpose of this example is to demonstrate how to export data from Stardog database in an obfuscated format.
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
		log.Fatalf("Error creating client: %v", err)
	}
	fmt.Print("Database to export default graph of: ")
	database, _ := r.ReadString('\n')
	database = strings.TrimSpace(database)

	obfuscationConfig, err := os.Open("obfuscation-config.ttl")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer obfuscationConfig.Close()

	opts := &stardog.ExportObfuscatedDataOptions{
		NamedGraph:        []string{"tag:stardog:api:context:default"},
		Format:            stardog.RDFFormatTurtle,
		ObfuscationConfig: obfuscationConfig,
	}

	buf, _, err := client.DatabaseAdmin.ExportObfuscatedData(context.Background(), database, opts)
	if err != nil {
		fmt.Println("unable to export data")
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}
	fmt.Printf("Successfully exported database: \"%s\"\n", database)
	fmt.Println("-------OBFUSCATED DATA-------")
	if buf != nil {
		fmt.Println(buf.String())
	}
}
