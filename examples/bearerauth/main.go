// The purpose of this example is to demonstrate how to create a client using bearer/token authentication.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("Endpoint: ")
	endpoint, _ := r.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)

	fmt.Print("Token: ")
	token, _ := r.ReadString('\n')
	fmt.Println()

	bearerAuthTransport := stardog.BearerAuthTransport{
		BearerToken: strings.TrimSpace(token),
	}

	client, err := stardog.NewClient(endpoint, bearerAuthTransport.Client())
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	isAlive, _, err := client.ServerAdmin.IsAlive(context.Background())
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
	fmt.Printf("Is %v alive?: %v\n", endpoint, *isAlive)
}
