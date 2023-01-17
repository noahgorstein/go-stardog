// The purpose of this example is to demonstrate how get the exact size of each database.
package main

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"

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
	client, _ := stardog.NewClient("http://localhost:5820", basicAuthTransport.Client())

	dbs, _, err := client.DatabaseAdmin.GetDatabases(context.Background())
	if err != nil {
		stardogErr, ok := err.(*stardog.ErrorResponse)
		if ok {
			fmt.Printf("HTTP Status: %v\n", stardogErr.Response.Status)
			fmt.Printf("Stardog Error Code: %v\n", stardogErr.Code)
			fmt.Printf("Stardog Error Message: %v\n", stardogErr.Message)
			os.Exit(1)
		}
		// some other error took place
		fmt.Println(err)
		os.Exit(1)
	}

	for _, db := range dbs {
		size, _, err := client.DatabaseAdmin.GetDatabaseSize(context.Background(), db, &stardog.GetDatabaseSizeOptions{Exact: true})
		if err != nil {
			stardogErr, ok := err.(*stardog.ErrorResponse)
			if ok {
				fmt.Printf("HTTP Status: %v\n", stardogErr.Response.Status)
				fmt.Printf("Stardog Error Code: %v\n", stardogErr.Code)
				fmt.Printf("Stardog Error Message: %v\n", stardogErr.Message)
				os.Exit(1)
			}
			// some other error took place
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Database: %s ---- Size: %d\n", db, *size)
	}
}
