// The purpose of this example is to demonstrate how get the exact size of each database.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {

	basicAuth := stardog.BasicAuthTransport{
		Username: "admin",
		Password: "admin",
	}
	client, _ := stardog.NewClient("http://localhost:5820", basicAuth.Client())

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
