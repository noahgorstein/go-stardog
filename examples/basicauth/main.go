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

	fmt.Print("Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	fmt.Println()

	basicAuthTransport := stardog.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}
	endpoint = strings.TrimSpace(endpoint)

	client, err := stardog.NewClient(endpoint, basicAuthTransport.Client())
	if err != nil {
		fmt.Println(fmt.Printf("Error creating client: %v", err))
		return
	}
	isAlive, _, err := client.ServerAdmin.IsAlive(context.Background())
	if err != nil {
		stardogErr, ok := err.(*stardog.ErrorResponse)
		if ok {
			fmt.Printf("HTTP Status: %v\n", stardogErr.Response.Status)
			fmt.Printf("Stardog Error Code: %v\n", stardogErr.Code)
			fmt.Printf("Stardog Error Message: %v\n", stardogErr.Message)
			return
		}
		// some other error took place
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("Is %v alive?: %v", endpoint, *isAlive))
}
