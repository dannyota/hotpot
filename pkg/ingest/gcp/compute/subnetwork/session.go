package subnetwork

import (
	"context"
	"sync"

	"google.golang.org/api/option"

	"hotpot/pkg/base/config"
)

// sessionClients stores clients keyed by Temporal session ID.
// Client lifetime = workflow session lifetime.
var sessionClients sync.Map

// GetOrCreateSessionClient returns existing client for session or creates new one.
func GetOrCreateSessionClient(ctx context.Context, sessionID string, configService *config.Service) (*Client, error) {
	if client, ok := sessionClients.Load(sessionID); ok {
		return client.(*Client), nil
	}

	// Create new client - prefer JSON credentials (from Vault) over file path
	var opts []option.ClientOption
	if credJSON := configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	} else if credFile := configService.GCPCredentialsFile(); credFile != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, credFile))
	}
	// If both empty, uses Application Default Credentials (ADC)

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
