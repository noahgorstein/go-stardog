
# go-stardog

[![Go Report Card](https://goreportcard.com/badge/github.com/noahgorstein/go-stardog)](https://goreportcard.com/report/github.com/noahgorstein/go-stardog)

go-stardog is a Go client library for interacting with a Stardog server.

## Example

Basic example of creating a Stardog client and listing the roles in Stardog.

```go
import (
	"context"
	"fmt"

	"github.com/noahgorstein/go-stardog/stardog"
)

func main() {
	client := stardog.NewClient("http://localhost:5820", "admin", "admin")
	roleList, _ := client.GetRoles(context.Background())
	fmt.Println(roleList.Roles)
}
```

## Notes

- This library is being actively worked on and is unstable. 
- This library is **not** officially maintained by Stardog.

## TODO

- Create a wrapper over all Stardog HTTP endpoints
- Add tests
