package main

import (
	"context"
	"fmt"
	"os"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {
	basicAuth := stardog.BasicAuthTransport{
		Username: "noah.gorstein@stardog.com",
		Password: "RinaBardin51!",
	}
	client, _ := stardog.NewClient("https://sd-deba99d4.stardog.cloud:5820", basicAuth.Client())

	dbOpts := map[string]interface{}{
		"spatial.enabled": true,
	}
	datasets := []stardog.Dataset{
		{
			Path:    "/Users/noahgorstein/projects/stardog-tutorials/music/beatles.ttl",
			Context: "http://my-graph",
		},
		{
			Path:    "/Users/noahgorstein/projects/stardog-tutorials/music/music_schema.ttl",
			Context: "http://graph-1",
		},
	}
	dbs, _, _ := client.DatabaseAdmin.GetDatabases(context.Background())
	fmt.Println(dbs)
	for _, db := range dbs {
		if db == os.Args[1] {
			fmt.Printf("dropping database: %s\n", os.Args[1])
			_, err := client.DatabaseAdmin.DropDatabase(context.Background(), os.Args[1])
			if err != nil {
				stardogErr, ok := err.(*stardog.ErrorResponse)
				if ok {
					fmt.Printf("HTTP Status: %v\n", stardogErr.Response.Status)
					fmt.Printf("Stardog Error Code: %v\n", stardogErr.Code)
					fmt.Printf("Stardog Error Message: %v\n", stardogErr.Message)
					os.Exit(1)
				}
				fmt.Println(err)
			}
		}
	}

	_, err := client.DatabaseAdmin.CreateDatabase(context.Background(), os.Args[1], datasets, dbOpts, false)
	if err != nil {
		stardogErr, ok := err.(*stardog.ErrorResponse)
		if ok {
			fmt.Printf("HTTP Status: %v\n", stardogErr.Response.Status)
			fmt.Printf("Stardog Error Code: %v\n", stardogErr.Code)
			fmt.Printf("Stardog Error Message: %v\n", stardogErr.Message)
			os.Exit(1)
		}
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("created database: %s\n", os.Args[1])
}
