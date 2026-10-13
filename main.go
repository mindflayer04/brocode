package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/mindflayer04/brocode/internal/openrouter"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "brocode:", err)
		os.Exit(1)
	}
}

func run() error {
	model := flag.String("model", "nvidia/nemotron-3-ultra-550b-a55b:free", "OpenRouter model ID")
	prompt := flag.String("prompt", "Say hello in one sentence.", "Message to send to the model")
	flag.Parse()

	if strings.TrimSpace(*model) == "" {
		return fmt.Errorf("choose a model with -model provider/model-name")
	}
	if strings.TrimSpace(*prompt) == "" {
		return fmt.Errorf("prompt must not be empty")
	}
	if flag.NArg() != 0 {
		return fmt.Errorf("unexpected arguments; use -prompt to supply your message")
	}

	client, err := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))
	if err != nil {
		return err
	}

	// Ctrl+C cancels the HTTP request.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	reply, err := client.Chat(ctx, *model, []openrouter.Message{
		{Role: "user", Content: *prompt},
	})
	if err != nil {
		return err
	}
	fmt.Println(reply)
	return nil
}
