// The purpose of this example is to demonstrate how to work with database metadata (a.k.a configuration options).
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

	fmt.Printf("Database to set 'search.enabled=true' for (%v): ", databases)
	database, _ := r.ReadString('\n')
	database = strings.TrimSpace(database)

	fmt.Printf("Offlining the database %s\n", database)
	_, err = client.DatabaseAdmin.OfflineDatabase(context.Background(), database)
	if err != nil {
		fmt.Printf("Unable to offline database \"%s\"\n", database)
		if !checkStardogError(err) {
			fmt.Println(err)
		}
	}
	fmt.Println("Database offlined successfully.")

	fmt.Println("Setting the database configuration option 'spatial.enabled=true'")
	setOptions := map[string]interface{}{
		"search.enabled": true,
	}
	_, err = client.DatabaseAdmin.SetDatabaseOptions(context.Background(), database, setOptions)
	if err != nil {
		fmt.Printf("Unable to set 'search.enabled' for database \"%s\"\n", database)
		if checkStardogError(err) {
			os.Exit(1)
		}
		// some other error took place
		fmt.Println(err)
		os.Exit(1)
	}

	configOptions := []string{"search.enabled"}
	opts, _, err := client.DatabaseAdmin.GetDatabaseOptions(context.Background(), database, configOptions)
	if err != nil {
		fmt.Printf("Unable to get value set for 'search.enabled' for database \"%s\"\n", database)
		if checkStardogError(err) {
			os.Exit(1)
		}
		// some other error took place
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("-----OPTIONS----")
	for key, value := range opts {
		fmt.Printf("OPTION: %s | VALUE: %v\n", key, value)
	}
	fmt.Println("----------------")

	fmt.Printf("Onlining the database %s\n", database)
	_, err = client.DatabaseAdmin.OnlineDatabase(context.Background(), database)
	if err != nil {
		fmt.Printf("Unable to online database \"%s\"\n", database)
		if !checkStardogError(err) {
			// some other error took place
			fmt.Println(err)
			os.Exit(1)
		}
	}
	fmt.Println("Database onlined successfully.")
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
