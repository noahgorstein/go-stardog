// The purpose of this example is to demonstrate how to create a client using bearer/token authentication.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
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
		log.Fatalf("unable to create Stardog client: %v", err)
	}
	isAlive, _, err := client.ServerAdmin.IsAlive(context.Background())
	if err != nil {
		var stardogErr *stardog.ErrorResponse
		if errors.As(err, &stardogErr) {
			log.Fatalf("stardog error occurred: %v", err)
		}
		log.Fatalf("non-stardog error occurred: %v", err)
	}
	fmt.Printf("Is %v alive?: %v\n", endpoint, *isAlive)
}
