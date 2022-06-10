
# go-stardog

![go-stardog](https://user-images.githubusercontent.com/23270779/175647530-ae5a8681-87a6-471d-a03a-4c700610483d.jpg)

go-stardog is a Go client library for interacting with a Stardog server.

## Usage

Construct a new Stardog client, then use the various services on the client to access different parts of the Stardog API. For example:

```go
client := stardog.NewClient("http://localhost:5820", "username", "password")
userPermissions, _ := client.Security.GetUserPermissions(context.Background(), "frodo")
```

The services of a client divide the API into logical chunks and mostly correspond to the structure of the Stardog API documentation at [https://stardog-union.github.io/http-docs/](https://stardog-union.github.io/http-docs/) .

> NOTE: Using the context package, one can easily pass cancelation signals and deadlines to various services of the client for handling a request. In case there is no context available, then `context.Background()` can be used as a starting point.

## Notes

- This library is being actively worked on and is unstable. 
- This library is **not** officially maintained by Stardog.

## TODO

- Create a wrapper over all Stardog HTTP endpoints
- Add tests
