package main

import (
	"context"
	"fmt"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {
	t := &stardog.BasicAuthTransport{
		Username: "admin",
		Password: "admin",
	}
	fmt.Println(t.Username)

	bearer := &stardog.BearerAuthTransport{
		Bearer: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhZG1pbiIsImF1ZCI6Imh0dHBzOi8vc3RhcmRvZy5jb20vMzY1OTNhNzYtNDIyZi00ZmM4LWJjZTUtYTQ5ZTc5NGRjZDI5IiwiaXNzIjoiaHR0cDovL3N0YXJkb2cuY29tL3YxIiwic3RhcmRvZy11c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNjcxMTE2ODEyLCJub25jZSI6Ijc5NzQxNDMwNDYwODgzOTc3MzQiLCJqdGkiOiIyN2JkZjg3OS03NzEzLTQyMGYtYWI4ZC1kYWFkOTU2MjA5YTEifQ.IT0dTy1bk_36-poHzD6rT1au3fSBbOVnWoOa-pe4Xu8",
	}
	fmt.Println(bearer.Bearer)

	client := stardog.NewClient(bearer.Client())
	roles, resp, err := client.Security.ListRoles(context.TODO())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resp.Status)
	for _, u := range roles.Roles {
		fmt.Println(u)
	}

	newPerm := stardog.NewPermission("write", "db", []string{"*"})
	isGranted, _, err := client.Security.GrantRolePermission(context.TODO(), "reader", *newPerm)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(fmt.Sprintf("Is granted: %v", isGranted))
	}

	rolePerms, _, err := client.Security.GetRolePermissions(context.TODO(), "reader")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resp.Status)
	for _, u := range rolePerms.Permissions {
		fmt.Println(u)
	}

  deleteRoleOptions := stardog.DeleteRoleOptions{
    Force: true,
  }
  didRoleGetDeleted, _, err := client.Security.DeleteRole(context.TODO(), "reader", &deleteRoleOptions)
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  fmt.Println(fmt.Sprintf("Deleted?: %v", didRoleGetDeleted))

}
