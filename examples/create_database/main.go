// The purpose of this example is to demonstrate how to create a database.
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

	datasets := []stardog.Dataset{
		{
			Path:       "./data/beatles.ttl",
			NamedGraph: "http://beatles",
		},
		{
			Path:       "./data/music.ttl.gz",
			NamedGraph: "http://music",
		},
		{
			Path:       "./data/music_schema.ttl",
			NamedGraph: "http://music",
		},
	}

	databaseOptions := map[string]interface{}{
		// enable search
		"search.enabled": true,
		// enabled named graph aliases
		"graph.aliases": true,
	}

	opts := &stardog.CreateDatabaseOptions{
		Datasets:        datasets,
		DatabaseOptions: databaseOptions,
		CopyToServer:    true,
	}

	msg, _, err := client.DatabaseAdmin.CreateDatabase(context.Background(), "go-stardog-test-db", opts)

	if err != nil {
		fmt.Println("Unable to create database")
		if checkStardogError(err) {
			os.Exit(1)
		}
		// some other error took place
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(*msg)

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
