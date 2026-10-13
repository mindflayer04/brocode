# Brocode

An LLM harness for software engineering, written in Go.

This is a minimal starting point. The executable sends a text prompt to a chosen
OpenRouter model and prints its response. Tool execution will come later.

## Project layout

```text
brocode/
├── go.mod              # Module name and Go version
├── main.go             # Entry point for the executable
├── internal/           # Application packages
│   ├── openrouter/     # Text chat client and HTTP contract tests
│   ├── sandbox/        # Placeholder for command execution
│   └── README.md       # Guidance on adding packages
├── .gitignore          # Local files and build output to ignore
└── README.md
```

`package main` and the `main()` function tell Go where the program starts.
For now, we can build here and move code into `internal/` as responsibilities
become clearer. No external dependencies are needed yet.

## Run

With `OPENROUTER_API_KEY` already set in your environment, run from the project
directory:

```sh
go run . -prompt "Explain Go interfaces with a small example."
```

The model defaults to `apodex/apodex-1.1-mini:free`. Use `-model provider/model-name`
to choose another model from the [OpenRouter model catalog](https://openrouter.ai/models).
The `-prompt` flag defaults to a short greeting request.

The program reads the API key using `os.Getenv("OPENROUTER_API_KEY")`. It does
not load `.env` files. Requests time out after two minutes, and Ctrl+C cancels
the current request.

## Use the client in Go

```go
client, err := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))
if err != nil {
    return err
}

reply, err := client.Chat(ctx, model, []openrouter.Message{
    {Role: "system", Content: "You are a software engineering assistant."},
    {Role: "user", Content: "Explain this function."},
})
```

Import `github.com/mindflayer04/brocode/internal/openrouter` to use this package
within Brocode. Supply a `context.Context` as `ctx` and the chosen model ID as
`model`; check `err` before using `reply`. Include previous messages in the slice
to continue a conversation—the client does not store history itself.

The client uses Go's standard library and the OpenRouter
[Chat Completions API](https://openrouter.ai/docs/api/api-reference/chat/create-a-chat-completion).
It supports non-streaming text responses for now. Streaming, tool calls, and
automatic retries can be added when the harness needs them.

## Build

```sh
go build -o bin/brocode .
./bin/brocode -prompt "Hello!"
```

## Format and check

```sh
gofmt -w main.go internal/openrouter/*.go
go vet ./...
go test ./...
```

The client tests use fake HTTP responses, so they need no API key or network
access and do not incur API charges. Tests live beside the code in `_test.go`
files.
