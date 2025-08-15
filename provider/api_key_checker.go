package provider

import (
	"context"
)

// todo add layer
type ApiKeyChecker interface {
	Check(ctx context.Context, apiKey string) error
}

type apiKeyChecker struct {
}

func NewApiKeyChecker() ApiKeyChecker {
	return &apiKeyChecker{}
}

func (c *apiKeyChecker) Check(ctx context.Context, apiKey string) error {
	// TODO implement
	return nil
}
