package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// A fake transport verifies HTTP requests without a real API key or network calls.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestChatRequest(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.String() != chatURL {
			t.Errorf("unexpected endpoint: %s %s", req.Method, req.URL)
		}
		if req.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing bearer authentication")
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Error("missing JSON content type")
		}
		var body chatRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "example/chosen-model" || body.Stream || len(body.Messages) != 2 {
			t.Fatalf("unexpected request: %+v", body)
		}
		if body.Messages[0].Role != "system" || body.Messages[1].Content != "Explain Go's \"interfaces\".\nBriefly." {
			t.Errorf("messages were not preserved: %+v", body.Messages)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"Hello!"},"finish_reason":"stop"}]}`)),
		}, nil
	})
	reply, err := client.Chat(context.Background(), "example/chosen-model", []Message{
		{Role: "system", Content: "Be concise."},
		{Role: "user", Content: "Explain Go's \"interfaces\".\nBriefly."},
	})
	if err != nil || reply != "Hello!" {
		t.Fatalf("reply = %q, error = %v", reply, err)
	}
}

func TestChatErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"unauthorized", 401, `{"error":{"message":"Invalid API key","code":401}}`, "HTTP 401: Invalid API key"},
		{"rate limit", 429, `{"error":{"message":"Rate limited"}}`, "HTTP 429: Rate limited"},
		{"non JSON error", 502, "upstream unavailable", "HTTP 502: Bad Gateway"},
		{"error with HTTP 200", 200, `{"error":{"message":"Provider failed"}}`, "Provider failed"},
		{"choice error", 200, `{"choices":[{"error":{"message":"Disconnected"},"finish_reason":"error"}]}`, "Disconnected"},
		{"invalid JSON", 200, "not JSON", "decode OpenRouter response"},
		{"no choices", 200, `{"choices":[]}`, "no choices"},
		{"no text", 200, `{"choices":[{"message":{"content":null}}]}`, "no text content"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := NewClient("test-key")
			client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			_, err := client.Chat(context.Background(), "example/model", []Message{{Role: "user", Content: "Hello"}})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestValidationAndCancellation(t *testing.T) {
	if _, err := NewClient("  "); err == nil {
		t.Fatal("expected missing API key error")
	}
	client, _ := NewClient("test-key")
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, req.Context().Err()
	})
	messages := []Message{{Role: "user", Content: "Hello"}}
	if _, err := client.Chat(context.Background(), "", messages); err == nil {
		t.Fatal("expected missing model error")
	}
	if _, err := client.Chat(context.Background(), "example/model", nil); err == nil {
		t.Fatal("expected missing messages error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Chat(ctx, "example/model", messages); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
