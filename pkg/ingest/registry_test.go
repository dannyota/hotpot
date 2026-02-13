package ingest

import (
	"io"
	"testing"

	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

func TestProviders_Empty(t *testing.T) {
	ResetProviders()
	defer ResetProviders()

	got := Providers()
	if len(got) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(got))
	}
}

func TestRegisterProvider(t *testing.T) {
	ResetProviders()
	defer ResetProviders()

	RegisterProvider(ProviderRegistration{
		Name:      "test-a",
		TaskQueue: "queue-a",
		Enabled:   func(*config.Service) bool { return true },
		RateLimitPerMinute: func(*config.Service) int { return 100 },
		Register: func(worker.Worker, *config.Service, *ent.Client) io.Closer { return nil },
	})
	RegisterProvider(ProviderRegistration{
		Name:      "test-b",
		TaskQueue: "queue-b",
		Enabled:   func(*config.Service) bool { return false },
		RateLimitPerMinute: func(*config.Service) int { return 200 },
		Register: func(worker.Worker, *config.Service, *ent.Client) io.Closer { return nil },
	})

	got := Providers()
	if len(got) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(got))
	}
	if got[0].Name != "test-a" {
		t.Errorf("expected first provider name %q, got %q", "test-a", got[0].Name)
	}
	if got[1].Name != "test-b" {
		t.Errorf("expected second provider name %q, got %q", "test-b", got[1].Name)
	}
}

func TestProviders_ReturnsCopy(t *testing.T) {
	ResetProviders()
	defer ResetProviders()

	RegisterProvider(ProviderRegistration{
		Name:      "original",
		TaskQueue: "queue",
		Enabled:   func(*config.Service) bool { return true },
		RateLimitPerMinute: func(*config.Service) int { return 100 },
		Register: func(worker.Worker, *config.Service, *ent.Client) io.Closer { return nil },
	})

	got := Providers()
	got[0].Name = "mutated"

	got2 := Providers()
	if got2[0].Name != "original" {
		t.Errorf("Providers() did not return a copy; mutation leaked")
	}
}
