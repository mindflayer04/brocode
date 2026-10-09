# Brocode

An LLM harness for software engineering, written in Go.

This is a minimal starting point. The executable currently prints a welcome
message; model integration and tool execution will come later.

## Project layout

```text
brocode/
├── go.mod              # Module name and Go version
├── main.go             # Entry point for the executable
├── internal/           # Future application packages
│   └── README.md       # Guidance on adding packages
├── .gitignore          # Local files and build output to ignore
└── README.md
```

`package main` and the `main()` function tell Go where the program starts.
For now, we can build here and move code into `internal/` as responsibilities
become clearer. No external dependencies are needed yet.

## Run

From the project directory:

```sh
go run .
```

## Build

```sh
go build -o bin/brocode .
./bin/brocode
```

## Format and check

```sh
gofmt -w main.go
go vet ./...
go test ./...
```

There are no tests yet. As we add behavior, its tests will live beside the code
in files ending in `_test.go`.
