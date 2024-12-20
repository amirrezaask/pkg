package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"
)

type Config struct {
	VaultAddress  string
	VaultRoleId   string
	VaultSecretId string
}

type Service struct {
	Client *vault.Client
}

func NewClient(c Config) (*Service, error) {
	client, err := vault.NewClient(&vault.Config{
		Address: c.VaultAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create Vault client: %v", err)
	}
	secretID := &auth.SecretID{FromString: c.VaultSecretId}
	appRoleAuth, err := auth.NewAppRoleAuth(
		c.VaultRoleId,
		secretID,
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to initialize AppRole auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return nil, err
	}
	if authInfo == nil {
		return nil, fmt.Errorf("No auth info was returned after login")
	}

	return &Service{Client: client}, nil
}

func (v *Service) GetSecrets(path string, serviceName string) (map[string]string, error) {
	secretsData, err := v.Client.KVv2(path).Get(context.Background(), serviceName)
	if err != nil {
		return nil, fmt.Errorf("Unable to read secrets: %w", err)
	}

	secrets := make(map[string]string)
	for key, value := range secretsData.Data {
		secrets[key] = value.(string)
	}

	return secrets, nil
}
