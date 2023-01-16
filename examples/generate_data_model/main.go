// The purpose of this example is to demonstrate how to generate a data model from a Stardog database.
package main

import (
	"bufio"
	"context"
	"fmt"
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
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	databases, _, err := client.DatabaseAdmin.GetDatabases(context.Background())
	if err != nil {
		fmt.Println("Unable to get databases")
		if checkStardogError(err) {
			os.Exit(1)
		}
		// some other error took place
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Database to generate database model for (%v): ", databases)
	database, _ := r.ReadString('\n')
	database = strings.TrimSpace(database)

	opts := &stardog.GenerateDataModelOptions{
		Reasoning: false,
		Output:    "text",
	}

	buf, _, err := client.DatabaseAdmin.GenerateDataModel(context.Background(), database, opts)
	if err != nil {
		fmt.Printf("Unable to create generate data model for \"%s\"\n", database)
		if checkStardogError(err) {
			os.Exit(1)
		}
		// some other error took place
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Successfully generated data model for: \"%s\"\n", database)
  fmt.Println("-------DATA MODEL-------")
	if buf != nil {
		fmt.Println(buf.String())
	}
}

func checkStardogError(err error) bool {
	stardogErr, ok := err.(*stardog.ErrorResponse)
	if ok {
		fmt.Printf("HTTP Status: %v\n", stardogErr.Response.Status)
		fmt.Printf("Stardog Error Code: %v\n", stardogErr.Code)
		fmt.Printf("Stardog Error Message: %v\n", stardogErr.Message)
		return true
	}
	return false
}
