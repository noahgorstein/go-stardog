// The purpose of this example is to demonstrate how to work with various security related functions.
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
	fmt.Print("Endpoint: ")
	endpoint, _ := r.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)

	fmt.Print("Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	fmt.Println()

	fmt.Print("Role to create: ")
	rolename, _ := r.ReadString('\n')
	rolename = strings.TrimSpace(rolename)

	basicAuthTransport := stardog.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client, err := stardog.NewClient(endpoint, basicAuthTransport.Client())
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// create the role
	_, err = client.Security.CreateRole(context.Background(), rolename)
	if err != nil {
		fmt.Printf("Unable to create role \"%s\"\n", rolename)
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
	fmt.Printf("Successfully created role \"%s\"\n", rolename)

	// grant read all permissions to the role created
	readAllPermissions := stardog.NewPermission(stardog.Read, stardog.AllResourceTypes, []string{"*"})
	_, err = client.Security.GrantRolePermission(context.Background(), rolename, *readAllPermissions)
	if err != nil {
		fmt.Printf("Unable to grant permission to role \"%s\"\n", rolename)
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
	fmt.Printf("Successfully granted permission [%v] [%v] [%v] to role \"%s\"\n",
		readAllPermissions.Action,
		readAllPermissions.ResourceType,
		strings.Join(readAllPermissions.Resource, ""),
		rolename)
	os.Exit(0)

}
