package lesson

import (
	"context"
	"testing"
)

func TestNewGeneratorForProviderAllProviders(t *testing.T) {
	ctx := context.Background()
	for _, p := range []string{"anthropic", "openai", "googleai", "deepseek", "llama"} {
		g, err := NewGeneratorForProvider(ctx, p, "fake-key-123")
		if err != nil {
			t.Fatalf("provider %q: unexpected error %v", p, err)
		}
		if g == nil {
			t.Fatalf("provider %q: nil generator", p)
		}
	}
}

func TestNewGeneratorForProviderUnknown(t *testing.T) {
	if _, err := NewGeneratorForProvider(context.Background(), "no-such-provider", "k"); err == nil {
		t.Fatal("unknown provider must error")
	}
}
