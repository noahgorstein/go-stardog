package main

import (
	"context"
	"fmt"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {
	client := stardog.NewClient("http://localhost:5820", "admin", "admin")
	roles, _ := client.Security.RolePermissions(context.Background(), "myRole")
	fmt.Println(roles)
}
