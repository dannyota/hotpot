package project

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// sessionClients stores clients keyed by Temporal session ID.
// Client lifetime = workflow session lifetime.
var sessionClients sync.Map

// GetOrCreateSessionClient returns existing client for session or creates new one.
func GetOrCreateSessionClient(ctx context.Context, sessionID string, configService *config.Service, limiter *rate.Limiter) (*Client, error) {
	if client, ok := sessionClients.Load(sessionID); ok {
		return client.(*Client), nil
	}

	// Create new client - use Vault JSON credentials, fall back to ADC
	var opts []option.ClientOption
	if credJSON := configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}

	// Rate limit via gRPC interceptor
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(limiter)),
	))

	client, err := NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	sessionClients.Store(sessionID, client)
	return client, nil
}

// CloseSessionClient closes and removes client for session.
func CloseSessionClient(sessionID string) {
	if client, ok := sessionClients.LoadAndDelete(sessionID); ok {
		client.(*Client).Close()
	}
}
