# go-stardog

![go-stardog](https://user-images.githubusercontent.com/23270779/175647530-ae5a8681-87a6-471d-a03a-4c700610483d.jpg)

![Coverage](https://img.shields.io/badge/Coverage-96.2%25-brightgreen)


go-stardog is a Go client library for interacting with a Stardog server.

## Usage

Usage:


```go
import "github.com/noahgorstein/go-stardog/stardog"
```

Construct a new Stardog client, then use the various services on the client to
access different parts of the Stardog API. For example:

```go
ctx := context.Background()

basicAuthTransport := stardog.BasicAuthTransport{
  Username: "admin",
  Password: "admin",
}

client, _ := stardog.NewClient("http://localhost:5820", basicAuthTransport.Client())

// list all users in the server
users, _, err := client.Security.GetUsers(ctx)
```

The services of a client divide the API into logical chunks and roughly correspond to structure of the [Stardog HTTP API documentation](https://stardog-union.github.io/http-docs/)

> **Note**<br>
> Using the https://godoc.org/context package, one can easily
> pass cancelation signals and deadlines to various services of the client for
> handling a request. In case there is no context available, then `context.Background()`
> can be used as a starting point.

For more sample code snippets, head over to the [examples](https://github.com/noahgorstein/go-stardog/tree/main/examples) directory.

## Authentication

The go-stardog library does not directly handle authentication. Instead, when
creating a new client, pass an `http.Client` that can handle authentication for
you.

### Basic Authentication

For users who wish to authenticate via username and password (HTTP Basic Authentication), use the `BasicAuthTransport`:

```go
func main() {

  ctx := context.Background()

  basicAuthTransport := stardog.BasicAuthTransport{
    Username: "admin",
    Password: "admin",
  }

  client, _ := stardog.NewClient("http://localhost:5820", basicAuthTransport.Client())

  // list all users in the server
  users, _, err := client.Security.GetUsers(ctx)
}
```

### Token Authentication

For users who wish to authenticate via an access token (Bearer Authentication), use the `BearerAuthTransport`:

```go
func main() {

  ctx := context.Background()

  bearerAuthTransport := stardog.BearerAuthTransport{
    BearerToken: "...token...",
  }

  client, _ := stardog.NewClient("http://localhost:5820", bearerAuthTransport.Client())

  // list all users in the server
  users, _, err := client.Security.GetUsers(ctx)
}
```

## Notes

- This library is being actively worked on and is unstable. 
- This library is **not** officially maintained by Stardog.
