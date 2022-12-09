module github.com/noahgorstein/go-github/example

go 1.18

require (
	github.com/noahgorstein/go-stardog v0.2.3
	golang.org/x/crypto v0.3.0
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/term v0.2.0 // indirect
)

replace github.com/noahgorstein/go-stardog v0.2.3 => ../../go-stardog
