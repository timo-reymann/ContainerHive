package secrets

import (
	"fmt"
)

// SecretResolver interface defines the method for resolving secrets
type SecretResolver interface {
	// Resolve takes a secret value and returns the resolved secret or an error
	// Implementations should only return an error if the value is invalid.
	// When it does not return an error and an empty string is returned, the secret is marked as not being
	// handled by this resolver.
	Resolve(value string) (resolvedValue string, err error)
}

// Secret represents a named secret
type Secret struct {
	Name  string
	Value string
}

var resolverOrder = []string{
	envVarResolver,
	plainTextResolver,
	vaultResolver,
}

var resolvers = map[string]SecretResolver{
	plainTextResolver: &PlainTextResolver{},
	envVarResolver:    &EnvVarResolver{},
	vaultResolver:     &VaultSecretResolver{},
}

// Resolve resolves a secret value using the registered resolvers in priority order
func Resolve(secretType, value string) (string, error) {
	if secretType != "" {
		resolver, ok := resolvers[secretType]
		if !ok {
			return "", fmt.Errorf("no resolver could handle secret of type %s", secretType)
		}
		return resolver.Resolve(value)
	}

	for _, rtype := range resolverOrder {
		resolver := resolvers[rtype]
		resolved, err := resolver.Resolve(value)
		if resolved != "" {
			return resolved, nil
		}

		if err != nil {
			return "", err
		}
	}

	// If no resolvers could handle this, return a generic error
	return "", fmt.Errorf("no resolver could handle secret of type %s with value: %s", secretType, value)
}
